// Package bundle turns a help archive into something addressable: a slug, a
// version, a table of contents.
package bundle

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/tobar-segais/tobar-segais/internal/archive"
)

// LatestVersion names the version of an archive that declares none. It is
// also the alias any product answers to, so /docs/<slug>/latest/ reaches an
// unversioned bundle by either route.
const LatestVersion = "latest"

// Slugs land in URLs, so they are checked rather than escaped: a bundle with
// a bad slug is skipped loudly instead of served at a surprising path.
var slugRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// ValidSlug reports whether s is safe to place in a path segment.
func ValidSlug(s string) bool { return slugRE.MatchString(s) }

// Identity is who a bundle is and where the answer came from. Source is kept
// because "why is this served at that URL" is the question operators ask.
type Identity struct {
	Slug    string
	Version string
	Title   string
	Aliases []string
	Hidden  bool
	// Priority orders the catalogue. Higher comes first, and bundles that
	// share a priority are ordered by title. It is unbounded and may be
	// negative: an archive nobody should land on first can sink itself
	// without every other archive having to be raised above it.
	Priority int
	// Copyright is a notice about this bundle's content. It comes from the
	// bundle itself, so it travels with the documentation rather than with
	// whoever happens to be hosting it, and it overrides the server-wide
	// notice for this bundle's pages.
	Copyright string
	Source    string // metadata | osgi | maven | filename
}

// Semver parses the version leniently: "1.16" is accepted as 1.16.0, which
// matters because most real bundles carry Maven versions rather than strict
// semver. Precedence is semver's, so 1.16.0-SNAPSHOT sorts before 1.16.0.
func (i Identity) Semver() (*semver.Version, bool) {
	if i.Version == "" {
		return nil, false
	}
	v, err := semver.NewVersion(i.Version)
	if err != nil {
		return nil, false
	}
	return v, true
}

// A filename is split into a name and a trailing version: everything up to
// the last separator that is followed by something version-shaped becomes the
// slug, and the rest the version. So "foo-1.0.0.zip" is foo at 1.0.0.
//
// The separator may be a hyphen or an underscore, and a "v" prefix on the
// version is accepted and dropped, because all three turn up in the wild.
var jarVersionRE = regexp.MustCompile(`^(.*?)[-_][vV]?(\d+(?:\.\d+)*(?:[-.+][0-9A-Za-z-.]+)?)$`)

// Identify resolves a bundle's identity.
//
// Nothing has to be declared. A bundle may say who it is in its navigation
// document, but where it does not, the answer is taken from an OSGi manifest,
// from Maven's pom.properties, or from the filename -- so "foo-1.0.0.zip" is
// foo at 1.0.0 with no metadata at all.
//
// Sources are consulted in that order and each fills only what is still
// missing, so a navigation document declaring a version but no slug keeps its
// version and takes the slug from the filename. Being able to state one
// without the other matters: the version is usually what a build knows, and
// the slug is usually what the filename already says.
//
// A metadata file beside the archive -- handbook-1.0.0.toml next to
// handbook-1.0.0.zip -- is read before any of them, so whatever it states
// wins. That is the point of it: it is how an archive you cannot rebuild gets
// the identity you need it to have.
func Identify(a *archive.Archive, filename string) (Identity, error) {
	// filepath, not path: this is a name on disk. On Windows path.Base
	// leaves the whole "C:\\dir\\foo-1.0.0.zip" intact, and the slug then
	// becomes the entire path with the separators mangled into hyphens.
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))

	var id Identity
	var from []string
	take := func(src string, c Identity) {
		used := false
		if id.Slug == "" && c.Slug != "" {
			id.Slug, used = c.Slug, true
		}
		if id.Version == "" && c.Version != "" {
			id.Version, used = c.Version, true
		}
		if id.Title == "" && c.Title != "" {
			id.Title = c.Title
		}
		if len(id.Aliases) == 0 && len(c.Aliases) > 0 {
			id.Aliases = c.Aliases
		}
		if c.Hidden {
			id.Hidden = true
		}
		if id.Copyright == "" && c.Copyright != "" {
			id.Copyright = c.Copyright
		}
		if id.Priority == 0 && c.Priority != 0 {
			id.Priority = c.Priority
		}
		if used {
			from = append(from, src)
		}
	}

	side, hidden, priority, ok, err := sidecarIdentity(filename)
	if err != nil {
		return Identity{}, fmt.Errorf("%s: %w", filepath.Base(SidecarPath(filename)), err)
	}
	if ok {
		take("sidecar", side)
	}
	if nav, ok := navIdentity(a); ok {
		take("nav", nav)
	}
	if id.Slug == "" || id.Version == "" {
		if name, version, ok := osgiIdentity(a); ok {
			take("osgi", Identity{Slug: normalise(name), Version: version})
		}
	}
	if id.Slug == "" || id.Version == "" {
		if name, version, ok := mavenIdentity(a); ok {
			take("maven", Identity{Slug: normalise(name), Version: version})
		}
	}
	if id.Slug == "" || id.Version == "" {
		if m := jarVersionRE.FindStringSubmatch(base); m != nil {
			take("filename", Identity{Slug: normalise(m[1]), Version: m[2]})
		} else {
			take("filename", Identity{Slug: normalise(base)})
		}
	}

	if id.Slug == "" {
		return Identity{}, fmt.Errorf("cannot determine a slug for %s", filepath.Base(filename))
	}
	// An archive that carries no version at all is still perfectly servable;
	// it just has one version, and "latest" is what to call it. Leaving this
	// empty would put an empty path segment in every URL the bundle appears
	// at, which is how it went wrong before.
	if id.Version == "" {
		id.Version = LatestVersion
		from = append(from, "default")
	}
	if !ValidSlug(id.Slug) {
		return Identity{}, fmt.Errorf("%q is not a valid slug", id.Slug)
	}
	// The metadata file has the last word here as well as the first: it is
	// the only source that can say "no, show this one", or return a bundle
	// that raised itself to the ordinary order of things.
	if hidden != nil {
		id.Hidden = *hidden
	}
	if priority != nil {
		id.Priority = *priority
	}
	id.Source = strings.Join(from, "+")
	return id, nil
}

// normalise makes a best effort at turning an arbitrary name into a valid
// slug rather than rejecting a bundle that is otherwise fine.
func normalise(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func osgiIdentity(a *archive.Archive) (name, version string, ok bool) {
	raw, err := a.ReadFile("META-INF/MANIFEST.MF")
	if err != nil {
		return "", "", false
	}
	h := parseManifest(raw)
	name = h["bundle-symbolicname"]
	// The header may carry directives, e.g. "com.example;singleton:=true".
	if i := strings.IndexByte(name, ';'); i != -1 {
		name = name[:i]
	}
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(h["bundle-version"])
	return name, version, name != ""
}

// parseManifest handles the jar manifest's 72-byte line folding, where a
// continuation begins with a single space.
func parseManifest(raw []byte) map[string]string {
	out := map[string]string{}
	var key string
	var val strings.Builder
	flush := func() {
		if key != "" {
			out[strings.ToLower(key)] = strings.TrimSpace(val.String())
		}
		key, val = "", strings.Builder{}
	}
	for _, line := range strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n") {
		if strings.HasPrefix(line, " ") {
			val.WriteString(line[1:])
			continue
		}
		flush()
		if i := strings.IndexByte(line, ':'); i != -1 {
			key = line[:i]
			val.WriteString(strings.TrimPrefix(line[i+1:], " "))
		}
	}
	flush()
	return out
}

func mavenIdentity(a *archive.Archive) (name, version string, ok bool) {
	for _, n := range a.Names() {
		if !strings.HasPrefix(n, "META-INF/maven/") || !strings.HasSuffix(n, "/pom.properties") {
			continue
		}
		raw, err := a.ReadFile(n)
		if err != nil {
			continue
		}
		props := map[string]string{}
		for _, line := range strings.Split(string(raw), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if i := strings.IndexByte(line, '='); i != -1 {
				props[strings.TrimSpace(line[:i])] = strings.TrimSpace(line[i+1:])
			}
		}
		if props["artifactId"] != "" {
			return props["artifactId"], props["version"], true
		}
	}
	return "", "", false
}
