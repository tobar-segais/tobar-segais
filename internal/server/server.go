// Package server serves the catalogue, bundle content and search.
package server

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/tobar-segais/tobar-segais/internal/bundle"
	"github.com/tobar-segais/tobar-segais/internal/search"
)

//go:embed templates/*.html static/*
var assets embed.FS

// Server ties the library and the index to HTTP.
type Server struct {
	lib *bundle.Library
	ix  *search.Indexer
	log *slog.Logger

	// defaultSlug, when set, makes "/" serve that product instead of the
	// catalogue -- for a server hosting one product, where an index listing
	// a single item is a page nobody wants to read.
	defaultSlug string
	// One template set per page. They cannot share a set: each page file
	// defines a block called "body", and parsing them together would leave
	// only the last one, silently rendering the wrong page inside the right
	// layout.
	pages map[string]*template.Template

	warnOnce sync.Once
}

// Options configures a server. Every field is optional.
type Options struct {
	// DefaultSlug makes "/" serve that product instead of the catalogue.
	DefaultSlug string
	// Site is the path the whole site is rooted at, empty when it is the
	// root of the domain.
	Site string
	// Copyright is a notice about the documentation being served, shown in
	// the footer. What it should say is the operator's business: they are
	// publishing someone's manuals, and this program has no way of knowing
	// whose.
	Copyright string
	// Static marks pages written to disk rather than served from memory.
	Static bool
}

// New builds the server and parses templates.
func New(lib *bundle.Library, ix *search.Indexer, o Options, log *slog.Logger) (*Server, error) {
	return newServer(lib, ix, o, log)
}

func newServer(lib *bundle.Library, ix *search.Indexer, o Options, log *slog.Logger) (*Server, error) {
	pages := map[string]*template.Template{}
	for _, name := range []string{"catalogue.html", "docs.html", "search.html", "search-static.html", "notfound.html"} {
		t, err := template.New(name).Funcs(funcs(o.Site, o.Copyright, o.Static)).
			ParseFS(assets, "templates/layout.html", "templates/sidebar.html", "templates/"+name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		pages[name] = t
	}
	return &Server{lib: lib, ix: ix, defaultSlug: o.DefaultSlug, log: log, pages: pages}, nil
}

// Handler returns the routed handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.catalogue)
	// Both spellings. A generated site has no routing, so the search page
	// is a file called search.html and that is what every page links to;
	// the served site answers the same address so the two behave alike.
	// /search stays for anything that already links to it.
	mux.HandleFunc("GET /search", s.search)
	mux.HandleFunc("GET /search.html", s.search)
	mux.HandleFunc("GET /docs/{slug}/{$}", s.productRoot)
	mux.HandleFunc("GET /docs/{slug}/{version}/{$}", s.versionRoot)
	mux.HandleFunc("GET /docs/{slug}/{version}/{path...}", s.content)
	mux.Handle("GET /static/", http.FileServerFS(assets))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok\n"))
	})
	return s.withLogging(mux)
}

// warnDefaultOnce complains about an unresolvable default exactly once, so a
// misconfiguration is visible without filling the log on every request.
func (s *Server) warnDefaultOnce() {
	s.warnOnce.Do(func() {
		s.log.Warn("default slug does not resolve, serving the catalogue instead",
			"slug", s.defaultSlug)
	})
}

func (s *Server) withLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		s.log.Debug("request", "method", r.Method, "path", r.URL.Path)
	})
}

// --- pages -----------------------------------------------------------------

type pageData struct {
	Title    string
	Products []*bundle.Product
	Product  *bundle.Product
	Bundle   *bundle.Bundle
	TOC      *bundle.TOC
	Current  string
	Content  template.HTML
	Headings []Heading
	Prev     *Link
	Next     *Link
	// Missing and Elsewhere describe a page that was asked for and is not
	// here, so a 404 inside a manual can still be useful.
	Missing   string
	Elsewhere []*bundle.Bundle
	// Versions are the switch-version links for the current page, resolved
	// against what each version actually contains.
	Versions []VersionLink
	Query    string
	Results  []search.Result
	Status   search.Status
	Base     string
}

func (s *Server) render(w http.ResponseWriter, name string, d pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.renderTo(w, name, d); err != nil {
		s.log.Error("template", "name", name, "err", err)
	}
}

// renderTo writes a page to any writer, so the same templates produce both
// the served site and the generated one.
func (s *Server) renderTo(w io.Writer, name string, d pageData) error {
	t, ok := s.pages[name]
	if !ok {
		return fmt.Errorf("no such template: %s", name)
	}
	return t.ExecuteTemplate(w, "layout", d)
}

func (s *Server) catalogue(w http.ResponseWriter, r *http.Request) {
	// A configured default takes the place of the index. If it does not
	// resolve -- misspelled, or its archive is not loaded yet -- fall back to
	// the catalogue rather than serving a redirect loop or a 404 at the root.
	if s.defaultSlug != "" {
		if p, ok := s.lib.Catalogue().Product(s.defaultSlug); ok {
			if b := p.Latest(); b != nil {
				http.Redirect(w, r, "/docs/"+p.Slug+"/"+b.Version+"/", http.StatusFound)
				return
			}
		}
		s.warnDefaultOnce()
	}
	s.render(w, "catalogue.html", pageData{
		Title:    "Documentation",
		Products: s.lib.Catalogue().Products,
		Status:   s.ix.Status(),
	})
}

// productRoot sends /docs/<slug>/ to the newest visible version.
func (s *Server) productRoot(w http.ResponseWriter, r *http.Request) {
	p, ok := s.lib.Catalogue().Product(r.PathValue("slug"))
	if !ok {
		s.notFound(w, r)
		return
	}
	b := p.Latest()
	if b == nil {
		s.notFound(w, r)
		return
	}
	http.Redirect(w, r, "/docs/"+p.Slug+"/"+b.Version+"/", http.StatusFound)
}

// versionRoot shows a version's landing page: its TOC root topic if it has
// one, otherwise the first topic with content.
func (s *Server) versionRoot(w http.ResponseWriter, r *http.Request) {
	b, ok := s.resolve(r)
	if !ok {
		s.notFound(w, r)
		return
	}
	href := b.TOC.Href
	if href == "" {
		b.TOC.Walk(func(t *bundle.Topic) {
			if href == "" && t.Href != "" {
				href = t.Href
			}
		})
	}
	if href == "" {
		s.notFound(w, r)
		return
	}
	http.Redirect(w, r, s.base(b)+href, http.StatusFound)
}

func (s *Server) base(b *bundle.Bundle) string {
	return "/docs/" + b.Slug + "/" + b.Version + "/"
}

// resolve finds the bundle for a request, accepting "latest" as a version.
func (s *Server) resolve(r *http.Request) (*bundle.Bundle, bool) {
	p, ok := s.lib.Catalogue().Product(r.PathValue("slug"))
	if !ok {
		return nil, false
	}
	v := r.PathValue("version")
	if v == "latest" {
		b := p.Latest()
		return b, b != nil
	}
	b := p.Find(v)
	return b, b != nil
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	s.render(w, "notfound.html", pageData{
		Title:    "Not found",
		Products: s.lib.Catalogue().Products,
	})
}

// notFoundIn reports a missing page from inside the manual it was asked for.
// The reader keeps the sidebar, the version pills and their place in the
// tree, which is the difference between a dead end and a wrong turn. If the
// page exists in another version, say so: that is the common case when a
// link is followed from an older release.
func (s *Server) notFoundIn(w http.ResponseWriter, r *http.Request, b *bundle.Bundle, missing string) {
	p := s.productOf(b)
	var elsewhere []*bundle.Bundle
	if p != nil {
		for _, other := range p.Versions {
			if other == b {
				continue
			}
			if _, ok := other.Archive.Lookup(missing); ok {
				elsewhere = append(elsewhere, other)
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
	s.render(w, "notfound.html", pageData{
		Title:     "Not found",
		Products:  s.lib.Catalogue().Products,
		Product:   p,
		Versions:  s.versionLinks(p, b, missing),
		Bundle:    b,
		TOC:       b.TOC,
		Base:      s.base(b),
		Current:   nearest(b.TOC, missing),
		Missing:   missing,
		Elsewhere: elsewhere,
	})
}

func isHTML(p string) bool {
	switch strings.ToLower(path.Ext(p)) {
	case ".html", ".htm", ".xhtml":
		return true
	}
	return false
}
