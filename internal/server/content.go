package server

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"github.com/tobar-segais/tobar-segais/internal/archive"
	"github.com/tobar-segais/tobar-segais/internal/bundle"
)

// content serves a file from inside a bundle. HTML pages are unwrapped and
// re-rendered inside the site layout, so navigation, search and versioning
// surround the documentation rather than the documentation being nailed
// inside a frame. Everything else is streamed straight out of the archive.
func (s *Server) content(w http.ResponseWriter, r *http.Request) {
	b, ok := s.resolve(r)
	if !ok {
		s.notFound(w, r)
		return
	}
	rel := r.PathValue("path")
	e, found := b.Archive.Lookup(rel)
	if !found {
		s.notFoundIn(w, r, b, rel)
		return
	}

	if !isHTML(rel) {
		s.serveRaw(w, r, b, e)
		return
	}

	d, err := s.docPage(b, rel)
	if err != nil {
		s.log.Error("read", "bundle", b.Slug, "path", rel, "err", err)
		http.Error(w, "cannot read page", http.StatusInternalServerError)
		return
	}
	s.render(w, "docs.html", d)
}

// docPage assembles everything a documentation page needs. Both the HTTP
// handler and the static site generator go through here, so a generated page
// is the same page.
func (s *Server) docPage(b *bundle.Bundle, rel string) (pageData, error) {
	raw, err := b.Archive.ReadFile(rel)
	if err != nil {
		return pageData{}, err
	}
	body, title, heads := bodyOf(raw)
	prev, next := neighbours(b.TOC, rel)
	if title == "" {
		title = labelFor(b.TOC, rel)
	}
	product := s.productOf(b)
	return pageData{
		Title:    title,
		Product:  product,
		Versions: s.versionLinks(product, b, rel),
		Bundle:   b,
		TOC:      b.TOC,
		Current:  rel,
		Content:  body,
		Headings: heads,
		Prev:     prev,
		Next:     next,
		Base:     s.base(b),
		Products: s.lib.Catalogue().Products,
		Status:   s.ix.Status(),
	}, nil
}

func (s *Server) productOf(b *bundle.Bundle) *bundle.Product {
	p, _ := s.lib.Catalogue().Product(b.Slug)
	return p
}

// serveRaw streams an asset. Where the client takes gzip and the entry is
// deflated, the compressed bytes go out untouched: no inflate, no
// recompress, just a copy from the archive to the socket.
func (s *Server) serveRaw(w http.ResponseWriter, r *http.Request, b *bundle.Bundle, e *archive.Entry) {
	h := w.Header()
	h.Set("ETag", e.ETag())
	h.Set("Content-Type", contentType(e.Name))
	if match := r.Header.Get("If-None-Match"); match != "" && strings.Contains(match, e.ETag()) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if !e.Stored() && acceptsGzip(r) {
		h.Set("Content-Encoding", "gzip")
		h.Set("Vary", "Accept-Encoding")
		h.Set("Content-Length", itoa(archive.GzipSize(e)))
		if r.Method == http.MethodHead {
			return
		}
		if err := b.Archive.GzipPassthrough(w, e); err != nil {
			s.log.Error("gzip passthrough", "path", e.Name, "err", err)
		}
		return
	}

	if e.Stored() {
		// Seekable, so ranges and conditional requests come for free.
		http.ServeContent(w, r, e.Name, e.Modified, b.Archive.Raw(e))
		return
	}
	rc, err := b.Archive.OpenEntry(e)
	if err != nil {
		http.Error(w, "cannot read", http.StatusInternalServerError)
		return
	}
	defer rc.Close()
	h.Set("Content-Length", itoa(e.Size))
	io.Copy(w, rc)
}

func acceptsGzip(r *http.Request) bool {
	for _, v := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.EqualFold(strings.TrimSpace(strings.SplitN(v, ";", 2)[0]), "gzip") {
			return true
		}
	}
	return false
}

func itoa(n int64) string {
	var b [20]byte
	i := len(b)
	if n == 0 {
		return "0"
	}
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func contentType(name string) string {
	switch strings.ToLower(name[strings.LastIndexByte(name, '.')+1:]) {
	case "css":
		return "text/css; charset=utf-8"
	case "js":
		return "application/javascript; charset=utf-8"
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "svg":
		return "image/svg+xml"
	case "pdf":
		return "application/pdf"
	case "json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

// Heading is an entry in the "On this page" column.
type Heading struct {
	ID    string
	Text  string
	Level int
}

// slugify turns heading text into an id stable enough to link to.
func slugify(s string) string {
	var b strings.Builder
	last := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			last = false
		default:
			if !last && b.Len() > 0 {
				b.WriteByte('-')
				last = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func textOf(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		if x.Type == html.TextNode {
			b.WriteString(x.Data)
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func setAttr(n *html.Node, key, val string) {
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

// normaliseHeadings shifts every heading so the shallowest one on the page is
// an h1.
//
// Generators disagree about what the top level means. Asciidoctor and goldmark
// make a document title an h1; Typst's HTML exporter reserves h1 for itself
// and starts content at h2; DocBook XSL emits a chapter title as h2. Rather
// than asking each format to compensate -- Typst cannot, since heading offsets
// may not be negative -- the server normalises on the way out, so a page reads
// the same whatever produced it.
func normaliseHeadings(root *html.Node) {
	level := func(n *html.Node) int {
		if n.Type != html.ElementNode || len(n.Data) != 2 || n.Data[0] != 'h' {
			return 0
		}
		if n.Data[1] < '1' || n.Data[1] > '6' {
			return 0
		}
		return int(n.Data[1] - '0')
	}

	min := 7
	var find func(*html.Node)
	find = func(n *html.Node) {
		if l := level(n); l > 0 && l < min {
			min = l
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}
	find(root)
	if min > 6 || min == 1 {
		return // nothing to do, or already correct
	}

	shift := min - 1
	var apply func(*html.Node)
	apply = func(n *html.Node) {
		if l := level(n); l > 0 {
			nl := l - shift
			if nl < 1 {
				nl = 1
			}
			n.Data = "h" + string(rune('0'+nl))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			apply(c)
		}
	}
	apply(root)
}

// collectHeadings finds headings for the page-level contents column,
// giving each an id if it has none so it can be linked to.
func collectHeadings(root *html.Node) []Heading {
	var out []Heading
	seen := map[string]int{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "h1" || n.Data == "h2" || n.Data == "h3") {
			text := textOf(n)
			if text != "" {
				id := attr(n, "id")
				if id == "" {
					id = slugify(text)
					if id == "" {
						id = "section"
					}
					if seen[id] > 0 {
						id = id + "-" + itoa(int64(seen[id]))
					}
					seen[id]++
					setAttr(n, "id", id)
				}
				out = append(out, Heading{ID: id, Text: text, Level: int(n.Data[1] - '0')})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	// The first heading is the page's own title, which the column sits beside
	// rather than lists.
	if len(out) > 0 && out[0].Level == 1 {
		out = out[1:]
	}
	return out
}

// bodyOf returns the inner HTML of <body> and the document title.
//
// Bundle content is authored HTML, not user input, so it is rendered as
// markup. Two things are removed: scripts, because a help page has no
// business running code in the container's origin, and stylesheets, because
// a page's CSS is written for the layout its generator produced and would
// restyle the one it is being rendered into.
//
// The packer removes stylesheets too, but that only covers bundles it built.
// An archive can be assembled by any means, so the check has to be here as
// well as there.
func bodyOf(raw []byte) (template.HTML, string, []Heading) {
	doc, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return template.HTML(template.HTMLEscapeString(string(raw))), "", nil
	}
	var body *html.Node
	var title string
	var find func(*html.Node)
	find = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "body":
				body = n
			case "title":
				if n.FirstChild != nil {
					title = strings.TrimSpace(n.FirstChild.Data)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}
	find(doc)
	if body == nil {
		return "", title, nil
	}
	stripScripts(body)
	stripChrome(body)
	normaliseHeadings(body)
	heads := collectHeadings(body)

	var sb strings.Builder
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(&sb, c); err != nil {
			break
		}
	}
	return template.HTML(sb.String()), title, heads
}

// generatorChrome is furniture a documentation generator bakes into each
// page, which the container provides for itself.
//
// DocBook XSL emits navheader/navfooter tables of Prev/Next links; leaving
// them in gives the reader two sets of navigation, one of them ugly.
// Asciidoctor emits a footer carrying a "Last updated" timestamp, which is
// noise inside a versioned archive.
var generatorChrome = map[string]bool{
	"navheader":   true, // DocBook XSL
	"navfooter":   true, // DocBook XSL
	"footer":      true, // Asciidoctor
	"footer-text": true, // Asciidoctor
}

// isChrome matches on class or id, because generators use both, and these
// names are specific enough to match at any depth.
func isChrome(n *html.Node) bool {
	for _, cls := range strings.Fields(attr(n, "class")) {
		if generatorChrome[cls] {
			return true
		}
	}
	return generatorChrome[attr(n, "id")]
}

// stripChrome removes generator furniture from a page body.
//
// Known ids and classes are matched anywhere. A semantic <footer> is matched
// only at the top level of the body, where it is the page's own footer rather
// than part of the content: a <footer> inside an article or a section belongs
// to whatever it sits in and is left alone.
//
// <header> is deliberately not removed at all. Generators put the document
// title there -- Asciidoctor in <div id="header">, Pandoc in
// <header id="title-block-header"> -- so dropping it would cost every page
// its heading.
func stripChrome(n *html.Node) { stripChromeAt(n, true) }

func stripChromeAt(n *html.Node, top bool) {
	var next *html.Node
	for c := n.FirstChild; c != nil; c = next {
		next = c.NextSibling
		if c.Type != html.ElementNode {
			continue
		}
		if isChrome(c) || (top && c.Data == "footer") {
			n.RemoveChild(c)
			continue
		}
		stripChromeAt(c, false)
	}
}

// unwanted reports elements that must not reach the rendered page.
func unwanted(n *html.Node) bool {
	switch n.Data {
	case "script", "noscript", "style":
		return true
	case "link":
		for _, a := range n.Attr {
			if strings.EqualFold(a.Key, "rel") &&
				strings.EqualFold(strings.TrimSpace(a.Val), "stylesheet") {
				return true
			}
		}
	case "base":
		// A <base> would silently re-point every relative link on the page.
		return true
	}
	return false
}

func stripScripts(n *html.Node) {
	var next *html.Node
	for c := n.FirstChild; c != nil; c = next {
		next = c.NextSibling
		if c.Type == html.ElementNode && unwanted(c) {
			n.RemoveChild(c)
			continue
		}
		if c.Type == html.ElementNode {
			attrs := c.Attr[:0]
			for _, a := range c.Attr {
				if strings.HasPrefix(strings.ToLower(a.Key), "on") {
					continue // inline event handlers
				}
				attrs = append(attrs, a)
			}
			c.Attr = attrs
		}
		stripScripts(c)
	}
}
