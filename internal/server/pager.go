package server

import (
	"strings"

	"github.com/tobar-segais/tobar-segais/internal/bundle"
)

// labelFor returns the table-of-contents label for a page. Content produced
// in a fragment mode -- Asciidoctor's -s, for instance -- carries no <title>,
// so the table of contents is the reliable source of a page's name.
func labelFor(toc *bundle.TOC, href string) string {
	var found string
	toc.Walk(func(t *bundle.Topic) {
		if found == "" && t.Href == href {
			found = t.Label
		}
	})
	return found
}

// nearest picks the entry in the table of contents closest to a page that is
// not there, so a 404 can still open the sidebar in roughly the right place.
// Closeness is the longest shared directory prefix: a missing
// "guides/install.html" settles on whatever else lives under "guides/".
func nearest(toc *bundle.TOC, missing string) string {
	want := strings.Split(strings.Trim(pathDir(missing), "/"), "/")
	best, bestScore := "", -1
	toc.Walk(func(t *bundle.Topic) {
		if t.Href == "" {
			return
		}
		have := strings.Split(strings.Trim(pathDir(t.Href), "/"), "/")
		n := 0
		for n < len(want) && n < len(have) && want[n] == have[n] && want[n] != "" {
			n++
		}
		if n > bestScore {
			best, bestScore = t.Href, n
		}
	})
	// Only claim a nearest page when there is a real shared prefix. A page
	// missing from the root of a manual has nothing in common with anything,
	// and marking the first entry current would tell the reader they are
	// somewhere they are not.
	if bestScore < 1 {
		return ""
	}
	return best
}

// pathDir is path.Dir without importing the whole of path for one call on a
// slash-separated archive path.
func pathDir(p string) string {
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[:i]
	}
	return ""
}

// Link is a neighbouring page in reading order.
type Link struct {
	Label string
	Href  string
}

// neighbours returns the pages either side of current in the table of
// contents. Reading order comes from the TOC itself, so it stays correct when
// a bundle is reorganised -- unlike the generator-baked Prev/Next links,
// which we strip.
func neighbours(toc *bundle.TOC, current string) (prev, next *Link) {
	var flat []Link
	toc.Walk(func(t *bundle.Topic) {
		if t.Href != "" {
			flat = append(flat, Link{Label: t.Label, Href: t.Href})
		}
	})
	for i, l := range flat {
		if l.Href != current {
			continue
		}
		if i > 0 {
			p := flat[i-1]
			prev = &p
		}
		if i+1 < len(flat) {
			n := flat[i+1]
			next = &n
		}
		return prev, next
	}
	return nil, nil
}
