package bundle

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/tobar-segais/tobar-segais/internal/archive"
)

// Bundle is one loaded archive: a single version of a single product.
type Bundle struct {
	Identity
	Path    string
	Archive *archive.Archive
	TOC     *TOC
}

// Title prefers the metadata title, then the TOC label, then the slug.
func (b *Bundle) Title() string {
	if b.Identity.Title != "" {
		return b.Identity.Title
	}
	if b.TOC != nil && b.TOC.Label != "" {
		return b.TOC.Label
	}
	return b.Slug
}

// Close releases every archive the library holds open.
//
// A server holds them for its lifetime and never calls this; a program that
// builds a site and stops should, and a test must -- on Windows an open file
// cannot be removed, so a library left open fails the cleanup of the very
// directory the test made.
func (l *Library) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, b := range l.loaded {
		b.Archive.Close()
		delete(l.loaded, key)
	}
	l.cat.Store(&Catalogue{bySlug: map[string]*Product{}})
}

// Product is every version of one slug, newest first.
type Product struct {
	Slug     string
	Versions []*Bundle
}

// Priority orders this product in the catalogue, and comes from its newest
// version: where a product sits is a property of what it is now, not of what
// it was three releases ago. A hidden-only product still has to sort, so it
// falls back to the newest version it has.
func (p *Product) Priority() int {
	if b := p.newest(); b != nil {
		return b.Priority
	}
	return 0
}

// Title is the newest version's title, for the same reason.
func (p *Product) Title() string {
	if b := p.newest(); b != nil {
		return b.Title()
	}
	return p.Slug
}

func (p *Product) newest() *Bundle {
	if b := p.Latest(); b != nil {
		return b
	}
	if len(p.Versions) > 0 {
		return p.Versions[0]
	}
	return nil
}

// Latest is the newest version that is not hidden.
func (p *Product) Latest() *Bundle {
	for _, b := range p.Versions {
		if !b.Hidden {
			return b
		}
	}
	return nil
}

// Find returns a specific version of this product.
func (p *Product) Find(version string) *Bundle {
	for _, b := range p.Versions {
		if b.Version == version {
			return b
		}
	}
	return nil
}

// Catalogue is an immutable snapshot of everything loaded. It is swapped
// wholesale so readers never see a half-built view and never take a lock.
type Catalogue struct {
	Products []*Product
	bySlug   map[string]*Product
}

// Product resolves a slug or one of its aliases.
func (c *Catalogue) Product(slug string) (*Product, bool) {
	p, ok := c.bySlug[slug]
	return p, ok
}

// Library owns the content directory and the current catalogue.
type Library struct {
	dir string
	log *slog.Logger

	mu     sync.Mutex // serialises reloads, not reads
	loaded map[string]*Bundle
	cat    atomic.Pointer[Catalogue]

	// OnChange is called after every successful reload, with the bundles
	// that are new since last time, so indexing can follow behind.
	OnChange func(added []*Bundle, cat *Catalogue)
}

// NewLibrary creates a library over dir. Nothing is read until Reload.
func NewLibrary(dir string, log *slog.Logger) *Library {
	l := &Library{dir: dir, log: log, loaded: map[string]*Bundle{}}
	l.cat.Store(&Catalogue{bySlug: map[string]*Product{}})
	return l
}

// Catalogue returns the current snapshot. Safe to call from any goroutine.
func (l *Library) Catalogue() *Catalogue { return l.cat.Load() }

// Dir is the watched content directory.
func (l *Library) Dir() string { return l.dir }

func isArchive(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".zip", ".jar":
		return true
	}
	return false
}

func isSidecar(name string) bool {
	return strings.EqualFold(filepath.Ext(name), SidecarExt)
}

// Reload rescans the directory, opening archives that are new and closing
// ones that have gone. Archives already loaded and unchanged are left alone,
// so a reload costs nothing for content that did not move.
func (l *Library) Reload() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	ents, err := os.ReadDir(l.dir)
	if err != nil {
		return err
	}

	seen := map[string]bool{}
	var added []*Bundle
	for _, ent := range ents {
		if ent.IsDir() || !isArchive(ent.Name()) {
			continue
		}
		p := filepath.Join(l.dir, ent.Name())
		info, err := ent.Info()
		if err != nil {
			continue
		}
		// Key on identity *and* mtime/size so a replaced file reloads. The
		// metadata file beside it counts too: editing it changes what the
		// bundle is, and an unchanged archive would otherwise be left alone.
		key := fmt.Sprintf("%s|%d|%d|%s", p, info.ModTime().UnixNano(), info.Size(), stamp(SidecarPath(p)))
		seen[key] = true
		if _, ok := l.loaded[key]; ok {
			continue
		}
		b, err := load(p)
		if err != nil {
			l.log.Warn("skipping archive", "path", p, "err", err)
			continue
		}
		l.log.Info("loaded bundle",
			"slug", b.Slug, "version", b.Version, "identity", b.Source, "path", p)
		l.loaded[key] = b
		added = append(added, b)
	}

	for key, b := range l.loaded {
		if !seen[key] {
			l.log.Info("removed bundle", "slug", b.Slug, "version", b.Version, "path", b.Path)
			b.Archive.Close()
			delete(l.loaded, key)
		}
	}

	cat := l.build()
	l.cat.Store(cat)
	if l.OnChange != nil {
		l.OnChange(added, cat)
	}
	return nil
}

// stamp identifies a file by when it changed and how big it is, and is empty
// when there is no file -- which is how adding or removing a metadata file
// reloads the archive beside it.
func stamp(p string) string {
	info, err := os.Stat(p)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d|%d", info.ModTime().UnixNano(), info.Size())
}

func load(p string) (*Bundle, error) {
	a, err := archive.Open(p)
	if err != nil {
		return nil, err
	}
	id, err := Identify(a, p)
	if err != nil {
		a.Close()
		return nil, err
	}
	toc, err := LoadTOC(a)
	if err != nil {
		a.Close()
		return nil, err
	}
	return &Bundle{Identity: id, Path: p, Archive: a, TOC: toc}, nil
}

// build groups the loaded bundles into products, newest version first.
func (l *Library) build() *Catalogue {
	bySlug := map[string]*Product{}
	for _, b := range l.loaded {
		p := bySlug[b.Slug]
		if p == nil {
			p = &Product{Slug: b.Slug}
			bySlug[b.Slug] = p
		}
		p.Versions = append(p.Versions, b)
	}

	for _, p := range bySlug {
		sort.SliceStable(p.Versions, func(i, j int) bool {
			return newer(p.Versions[i], p.Versions[j])
		})
	}

	// Aliases resolve to the same product, but never shadow a real slug.
	for _, p := range bySlug {
		for _, b := range p.Versions {
			for _, al := range b.Aliases {
				if _, taken := bySlug[al]; !taken {
					bySlug[al] = p
				}
			}
		}
	}

	products := make([]*Product, 0, len(bySlug))
	for slug, p := range bySlug {
		if slug == p.Slug {
			products = append(products, p)
		}
	}
	// Priority first, then title, then slug. Alphabetical order is a fine
	// default and a poor answer for a site with a landing page: the page a
	// reader should meet first is rarely the one whose title starts with A.
	// The priority of a product is the priority of its newest version, so
	// raising a manual is done by the release that raises it.
	sort.Slice(products, func(i, j int) bool {
		a, b := products[i], products[j]
		if pa, pb := a.Priority(), b.Priority(); pa != pb {
			return pa > pb
		}
		if ta, tb := a.Title(), b.Title(); ta != tb {
			return ta < tb
		}
		return a.Slug < b.Slug
	})
	return &Catalogue{Products: products, bySlug: bySlug}
}

// newer reports whether a sorts before b, i.e. is the more recent version.
// Semver precedence where both parse, so 1.16.0-SNAPSHOT sorts after 1.16.0
// in this ordering (it is older). Unparseable versions fall back to a string
// compare and always sort last, so they cannot become "latest" by accident.
func newer(a, b *Bundle) bool {
	av, aok := a.Semver()
	bv, bok := b.Semver()
	switch {
	case aok && bok:
		return bv.LessThan(av)
	case aok:
		return true
	case bok:
		return false
	default:
		return a.Version > b.Version
	}
}
