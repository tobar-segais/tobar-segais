package bundle

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// SidecarExt is the extension of the metadata file that may sit beside an
// archive: content/handbook-1.0.0.zip is described by
// content/handbook-1.0.0.toml.
const SidecarExt = ".toml"

// SidecarPath is where an archive's metadata file would be, whether or not
// one exists.
func SidecarPath(archivePath string) string {
	return strings.TrimSuffix(archivePath, filepath.Ext(archivePath)) + SidecarExt
}

// sidecar is the file's shape. Every field is optional: a metadata file that
// sets one thing says nothing about the rest.
type sidecar struct {
	Slug      string   `toml:"slug"`
	Version   string   `toml:"version"`
	Title     string   `toml:"title"`
	Aliases   []string `toml:"aliases"`
	Hidden    *bool    `toml:"hidden"`
	Priority  *int     `toml:"priority"`
	Copyright string   `toml:"copyright"`
}

// sidecarIdentity reads the metadata file beside an archive.
//
// This is the one source that can contradict the archive rather than merely
// fill gaps in it, which is what makes it useful: it is the only way to fix
// the identity of a bundle you did not build and cannot rebuild. A vendor's
// jar that names itself badly, a version that has to be corrected after
// release, an old bundle that predates any of this -- all of them are files
// you can put a text file next to, and none of them are files you can edit.
//
// A malformed file is an error rather than a shrug: someone wrote it meaning
// to change something, and quietly serving the archive's own identity instead
// would hide that it did nothing.
// hidden and priority are returned separately because they are the fields
// where "not said" and "said the zero value" differ: a metadata file must be
// able to unhide a bundle that hid itself, and to put a bundle that raised
// itself back among the rest.
func sidecarIdentity(archivePath string) (id Identity, hidden *bool, priority *int, ok bool, err error) {
	raw, err := os.ReadFile(SidecarPath(archivePath))
	if err != nil {
		return Identity{}, nil, nil, false, nil // no metadata file is the normal case
	}
	var s sidecar
	if err := toml.Unmarshal(raw, &s); err != nil {
		return Identity{}, nil, nil, false, err
	}
	id = Identity{
		Slug:      normalise(s.Slug),
		Version:   strings.TrimSpace(s.Version),
		Title:     strings.TrimSpace(s.Title),
		Copyright: strings.TrimSpace(s.Copyright),
		Source:    "sidecar",
	}
	for _, a := range s.Aliases {
		if x := normalise(a); x != "" {
			id.Aliases = append(id.Aliases, x)
		}
	}
	if s.Hidden != nil {
		id.Hidden = *s.Hidden
	}
	if s.Priority != nil {
		id.Priority = *s.Priority
	}
	empty := id.Slug == "" && id.Version == "" && id.Title == "" &&
		id.Copyright == "" && len(id.Aliases) == 0 && s.Hidden == nil && s.Priority == nil
	return id, s.Hidden, s.Priority, !empty, nil
}
