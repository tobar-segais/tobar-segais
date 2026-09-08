package pack

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

// generatorChrome is page furniture the server supplies for itself, removed
// here so it is not carried in the archive at all. Asciidoctor's footer holds
// a "Last updated" timestamp, which would otherwise change every page on
// every build even when nothing had.
var generatorChrome = map[string]bool{
	"footer":      true,
	"footer-text": true,
	"navheader":   true,
	"navfooter":   true,
}

// stripStyles removes stylesheets and generator furniture from generated
// HTML.
//
// The server renders page content inside its own layout and never uses a
// bundle's head, so any CSS a generator emits is dead weight -- Asciidoctor
// embeds around 29 KB of it in every full document. Removing it here means
// converters can be invoked plainly instead of each needing the right flag to
// suppress something we were going to discard anyway, and it works for output
// we did not generate, such as a directory packed with --format html.
//
// Files carrying no stylesheet are left untouched, byte for byte, so this
// costs nothing for the formats that never emit any.
func stripStyles(dir string) error {
	return filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		switch strings.ToLower(filepath.Ext(p)) {
		case ".html", ".htm":
		default:
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if !hasStyles(raw) {
			return nil
		}
		out, err := removeStyles(raw)
		if err != nil {
			return nil // leave anything we cannot parse exactly as it was
		}
		return os.WriteFile(p, out, info.Mode().Perm())
	})
}

func hasStyles(raw []byte) bool {
	lower := bytes.ToLower(raw)
	if bytes.Contains(lower, []byte("<style")) ||
		bytes.Contains(lower, []byte(`rel="stylesheet"`)) ||
		bytes.Contains(lower, []byte("rel='stylesheet'")) {
		return true
	}
	if bytes.Contains(lower, []byte("<footer")) {
		return true
	}
	for name := range generatorChrome {
		if bytes.Contains(lower, []byte(`id="`+name+`"`)) ||
			bytes.Contains(lower, []byte(`class="`+name+`"`)) {
			return true
		}
	}
	return false
}

func removeStyles(raw []byte) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	// A semantic <footer> counts as chrome only directly inside <body>; one
	// nested in content belongs to the content.
	var walk func(n *html.Node, inBody bool)
	walk = func(n *html.Node, inBody bool) {
		var next *html.Node
		for c := n.FirstChild; c != nil; c = next {
			next = c.NextSibling
			if c.Type != html.ElementNode {
				continue
			}
			if drop(c) || (inBody && c.Data == "footer") {
				n.RemoveChild(c)
				continue
			}
			walk(c, c.Data == "body")
		}
	}
	walk(doc, false)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drop(n *html.Node) bool {
	if n.Data == "style" {
		return true
	}
	for _, a := range n.Attr {
		if (a.Key == "id" || a.Key == "class") && generatorChrome[strings.TrimSpace(a.Val)] {
			return true
		}
	}
	if n.Data != "link" {
		return false
	}
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, "rel") && strings.EqualFold(strings.TrimSpace(a.Val), "stylesheet") {
			return true
		}
	}
	return false
}
