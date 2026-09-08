package server

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/tobar-segais/tobar-segais/internal/bundle"
	"github.com/tobar-segais/tobar-segais/internal/search"
)

// search answers a query against the published index. If indexing is still
// running the page says so, rather than showing an empty result set that
// looks like an answer.
func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	slug := r.URL.Query().Get("slug")
	version := r.URL.Query().Get("version")

	var results []search.Result
	if q != "" {
		for _, hit := range s.ix.Index().Search(q, 100) {
			if slug != "" && hit.Slug != slug {
				continue
			}
			if version != "" && hit.Version != version {
				continue
			}
			results = append(results, hit)
			if len(results) == 50 {
				break
			}
		}
	}
	s.render(w, "search.html", pageData{
		Title:    "Search",
		Query:    q,
		Results:  results,
		Status:   s.ix.Status(),
		Products: s.lib.Catalogue().Products,
	})
}

func funcs(site, contentCopyright string, static bool) template.FuncMap {
	return template.FuncMap{
		// A running server serves; a generated site was published once and
		// is now just files. The footer says which, because it changes what
		// a reader can expect -- a served page is as current as the content
		// directory, a published one is as current as the last build.
		"serving": func() bool { return !static },
		"provenance": func() string {
			if static {
				return "Published by"
			}
			return "Served by"
		},
		// notice is what the footer says about the documentation on this
		// page. A bundle that carries its own notice wins: it travels with
		// the content, and the content is what the notice is about. The
		// operator's --copyright is the fallback, for pages that belong to
		// no one bundle and for archives that say nothing. Both are empty
		// unless somebody said something -- this program claims nothing
		// about documentation it did not write. The notice for the software
		// itself is separate, and is not either party's to change.
		"notice": func(b *bundle.Bundle) string {
			if b != nil && b.Copyright != "" {
				return b.Copyright
			}
			return contentCopyright
		},
		// site is the path the whole site is rooted at: empty when served
		// by this server, and possibly a subpath for a generated site.
		"site": func() string { return site },
		// highlight escapes the snippet, then turns the indexer's private
		// markers into emphasis. Escaping first means document text can
		// never inject markup.
		"highlight": func(s string) template.HTML {
			esc := template.HTMLEscapeString(s)
			esc = strings.ReplaceAll(esc, template.HTMLEscapeString(search.MarkStart), "<mark>")
			esc = strings.ReplaceAll(esc, template.HTMLEscapeString(search.MarkEnd), "</mark>")
			return template.HTML(esc)
		},
		// dict lets the recursive topic template carry context down the
		// tree, which Go templates otherwise make awkward.
		"dict": func(kv ...any) map[string]any {
			m := make(map[string]any, len(kv)/2)
			for i := 0; i+1 < len(kv); i += 2 {
				k, _ := kv[i].(string)
				m[k] = kv[i+1]
			}
			return m
		},
		"pct": func(st search.Status) int {
			if st.Total == 0 {
				return 0
			}
			return int(st.Done * 100 / st.Total)
		},
	}
}
