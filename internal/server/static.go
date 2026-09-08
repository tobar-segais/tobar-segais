package server

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/tobar-segais/tobar-segais/internal/bundle"
	"github.com/tobar-segais/tobar-segais/internal/search"
)

// StaticOptions configures a generated site.
type StaticOptions struct {
	Out string
	// Site is the path the site is rooted at, for hosting under a subpath.
	// A GitHub Pages project site lives at /<repo>/, for instance.
	Site string
	// DefaultSlug, when set, makes the root page the newest version of that
	// product rather than the catalogue.
	DefaultSlug string
	// Copyright is a notice about the documentation, shown in the footer.
	Copyright string
}

// GenerateStatic writes the whole site to disk as plain files.
//
// Every page is rendered with the same templates the server uses, so what is
// generated is what would have been served. What cannot be static is search:
// there is no server to ask, so the index is shipped as JSON and queried in
// the browser.
func GenerateStatic(lib *bundle.Library, ix *search.Indexer, o StaticOptions, log *slog.Logger) error {
	site := strings.TrimSuffix(o.Site, "/")
	s, err := newServer(lib, ix, Options{DefaultSlug: o.DefaultSlug, Site: site, Copyright: o.Copyright, Static: true}, log)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(o.Out, 0o755); err != nil {
		return err
	}

	cat := lib.Catalogue()
	var pages, assets int

	// The server's own stylesheet, logo and icons.
	if err := copyEmbedded(o.Out); err != nil {
		return err
	}

	for _, p := range cat.Products {
		latest := p.Latest()
		if latest == nil {
			continue
		}
		// /docs/<slug>/ goes to the newest version.
		if err := writeRedirect(
			filepath.Join(o.Out, "docs", p.Slug, "index.html"),
			site+"/docs/"+p.Slug+"/"+latest.Version+"/"); err != nil {
			return err
		}

		for _, b := range p.Versions {
			n, a, err := writeVersion(s, o, site, b)
			if err != nil {
				return fmt.Errorf("%s %s: %w", b.Slug, b.Version, err)
			}
			pages += n
			assets += a
		}

		// "latest" is an alias the server resolves at request time; a static
		// site needs something at that address, so every page of the newest
		// version gets a redirect stub there.
		for _, href := range pageHrefs(latest) {
			if err := writeRedirect(
				filepath.Join(o.Out, "docs", p.Slug, "latest", filepath.FromSlash(href)),
				site+"/docs/"+p.Slug+"/"+latest.Version+"/"+href); err != nil {
				return err
			}
		}
		if err := writeRedirect(
			filepath.Join(o.Out, "docs", p.Slug, "latest", "index.html"),
			site+"/docs/"+p.Slug+"/"+latest.Version+"/"); err != nil {
			return err
		}
	}

	// Root page.
	rootData := pageData{Title: "Documentation", Products: cat.Products, Status: ix.Status()}
	if o.DefaultSlug != "" {
		if p, ok := cat.Product(o.DefaultSlug); ok {
			if b := p.Latest(); b != nil {
				if err := writeRedirect(filepath.Join(o.Out, "index.html"),
					site+"/docs/"+p.Slug+"/"+b.Version+"/"); err != nil {
					return err
				}
				goto searchPage
			}
		}
		log.Warn("default slug does not resolve, generating the catalogue at the root",
			"slug", o.DefaultSlug)
	}
	if err := writeRendered(s, filepath.Join(o.Out, "index.html"), "catalogue.html", rootData); err != nil {
		return err
	}

searchPage:
	if err := writeRendered(s, filepath.Join(o.Out, "404.html"), "notfound.html",
		pageData{Title: "Not found", Products: cat.Products}); err != nil {
		return err
	}
	if err := writeSearch(s, o.Out, cat.Products, ix); err != nil {
		return err
	}

	log.Info("site generated", "out", o.Out, "pages", pages, "assets", assets,
		"documents", ix.Index().Len())
	return nil
}

// pageHrefs lists the pages of a bundle, in table-of-contents order,
// including the root topic.
func pageHrefs(b *bundle.Bundle) []string { return b.TOC.Pages() }

// writeVersion renders every page of one version and copies its assets.
func writeVersion(s *Server, o StaticOptions, site string, b *bundle.Bundle) (pages, assets int, err error) {
	root := filepath.Join(o.Out, "docs", b.Slug, b.Version)
	hrefs := pageHrefs(b)

	for _, href := range hrefs {
		d, err := s.docPage(b, href)
		if err != nil {
			// A table of contents may name a page the archive does not
			// contain; the server 404s, and here we simply skip it.
			continue
		}
		if err := writeRendered(s, filepath.Join(root, filepath.FromSlash(href)), "docs.html", d); err != nil {
			return pages, assets, err
		}
		pages++
	}

	// Anything in the archive that is not a page: images, downloads, CSS
	// referenced by content.
	for _, name := range b.Archive.Names() {
		if isHTML(name) || name == "nav.html" {
			continue
		}
		if strings.HasPrefix(name, "META-INF/") || name == "plugin.xml" {
			continue
		}
		raw, err := b.Archive.ReadFile(name)
		if err != nil {
			continue
		}
		dst := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return pages, assets, err
		}
		if err := os.WriteFile(dst, raw, 0o644); err != nil {
			return pages, assets, err
		}
		assets++
	}

	// The version's own landing page.
	landing := b.TOC.Href
	if landing == "" && len(hrefs) > 0 {
		landing = hrefs[0]
	}
	if landing != "" && landing != "index.html" {
		if err := writeRedirect(filepath.Join(root, "index.html"),
			site+"/docs/"+b.Slug+"/"+b.Version+"/"+landing); err != nil {
			return pages, assets, err
		}
	}
	return pages, assets, nil
}

func writeRendered(s *Server, dst, tmpl string, d pageData) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	return s.renderTo(f, tmpl, d)
}

// writeRedirect writes a page that sends a browser elsewhere. Static hosting
// has no redirects of its own, so this is how an alias is expressed.
func writeRedirect(dst, to string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	esc := html.EscapeString(to)
	body := "<!doctype html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n" +
		"<link rel=\"canonical\" href=\"" + esc + "\">\n" +
		"<meta http-equiv=\"refresh\" content=\"0; url=" + esc + "\">\n" +
		"<title>Redirecting</title>\n</head>\n<body>\n" +
		"<p>Redirecting to <a href=\"" + esc + "\">" + esc + "</a>.</p>\n</body>\n</html>\n"
	return os.WriteFile(dst, []byte(body), 0o644)
}

// copyEmbedded writes the server's own assets into the generated site.
func copyEmbedded(out string) error {
	entries, err := assets.ReadDir("static")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		raw, err := assets.ReadFile(path.Join("static", e.Name()))
		if err != nil {
			return err
		}
		dst := filepath.Join(out, "static", e.Name())
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, raw, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// writeSearch ships the index as JSON alongside a page that queries it in the
// browser.
func writeSearch(s *Server, out string, products []*bundle.Product, ix *search.Indexer) error {
	docs := ix.Index().Export()
	raw, err := json.Marshal(docs)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "search-index.json"), raw, 0o644); err != nil {
		return err
	}
	return writeRendered(s, filepath.Join(out, "search.html"), "search-static.html",
		pageData{Title: "Search", Products: products})
}

var _ = io.Discard
