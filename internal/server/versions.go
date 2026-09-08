package server

import (
	"strings"

	"github.com/tobar-segais/tobar-segais/internal/bundle"
)

// VersionLink is one entry in the version picker.
type VersionLink struct {
	Version  string
	Href     string
	IsHere   bool
	IsLatest bool
	// HasPage is false when this version does not contain the page being
	// read, in which case Href goes to the version instead.
	HasPage bool
}

// versionLinks builds the version picker for a page.
//
// Switching version is only useful if it lands somewhere. Versions differ in
// what they contain -- 2.0.0 of this manual has no copyright page, 1.15 has
// no page on source formats -- so a link built blind sends the reader to a
// page that was never written. Where the page is absent the link goes to the
// version's own landing page instead, and the control says so.
//
// The server can afford to answer that with a helpful 404; a generated site
// cannot, because the file simply is not there. Resolving the link here is
// what makes the two behave alike.
func (s *Server) versionLinks(p *bundle.Product, here *bundle.Bundle, current string) []VersionLink {
	if p == nil {
		return nil
	}
	latest := p.Latest()
	clean := current
	if i := strings.IndexAny(clean, "#?"); i != -1 {
		clean = clean[:i]
	}

	out := make([]VersionLink, 0, len(p.Versions))
	for _, v := range p.Versions {
		l := VersionLink{
			Version:  v.Version,
			Href:     s.base(v),
			IsHere:   here != nil && v.Version == here.Version,
			IsLatest: latest != nil && v.Version == latest.Version,
		}
		if clean != "" {
			if _, ok := v.Archive.Lookup(clean); ok {
				l.HasPage = true
				l.Href = s.base(v) + current
			}
		}
		out = append(out, l)
	}
	return out
}
