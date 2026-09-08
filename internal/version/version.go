// Package version reports the version that `tobar-segais version` prints.
package version

import (
	"runtime/debug"
	"strings"
)

// The version to report.
//
// THE REPOSITORY HOLDS NO NUMBER. The release workflow stamps this variable
// with `-ldflags "-X github.com/tobar-segais/tobar-segais/internal/version.version=X.Y.Z"`,
// and the number comes from the tag that the build builds. A build with no
// stamp asks the module system instead: `go install
// github.com/tobar-segais/tobar-segais@vX.Y.Z` records the version of the
// module, and a build from a checkout records the hash of the commit.
var version = ""

// String returns the version to report.
func String() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	// `go install module@vX.Y.Z` records the version of the module. A build
	// from a checkout records `(devel)`, which names nothing; the commit hash
	// below says more.
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return strings.TrimPrefix(v, "v")
	}

	// A build from a checkout: the hash of the commit, and `-dirty` when the
	// working tree held changes that the hash does not name.
	var revision, dirty string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				dirty = "-dirty"
			}
		}
	}
	if revision == "" {
		return "unknown"
	}
	if len(revision) > 12 {
		revision = revision[:12]
	}
	return revision + dirty
}
