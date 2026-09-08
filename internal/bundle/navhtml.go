package bundle

import (
	"bytes"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/tobar-segais/tobar-segais/internal/archive"
)

// NavFiles are the generated navigation documents we recognise, in order of
// preference. An HTML nav is what Asciidoctor produces from a nested list of
// xrefs, which means a bundle's navigation and its identity can both be
// generated from source rather than maintained alongside it.
var NavFiles = []string{"nav.html", "toc.html"}

// metaPrefix namespaces our <meta> names so they cannot collide with anything
// a documentation toolchain emits of its own accord.
const metaPrefix = "tobar-segais."

// ParseNavHTML reads a table of contents from a generated HTML navigation
// document: nested <ul> lists where each <li> holds a link, or plain text for
// a heading that groups its children.
func ParseNavHTML(raw []byte) (*TOC, error) {
	doc, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	list := firstNavList(doc)
	toc := &TOC{Label: strings.TrimSpace(titleOf(doc))}
	if list != nil {
		toc.Topics = topicsFrom(list)
	}
	return toc, nil
}

// firstNavList finds the outermost <ul>, skipping any inside the document
// header or a table of contents block the generator added for itself.
func firstNavList(n *html.Node) *html.Node {
	var found *html.Node
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		if found != nil {
			return
		}
		if x.Type == html.ElementNode && x.Data == "ul" {
			found = x
			return
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return found
}

func topicsFrom(ul *html.Node) []*Topic {
	var out []*Topic
	for li := ul.FirstChild; li != nil; li = li.NextSibling {
		if li.Type != html.ElementNode || li.Data != "li" {
			continue
		}
		t := &Topic{}
		// The link and label come from this item, not from its children:
		// Asciidoctor wraps item text in <p>, and nests sub-lists in <ul>.
		if a := firstLinkOutsideLists(li); a != nil {
			t.Href = hrefOf(a)
			t.Label = strings.TrimSpace(textExcludingLists(a))
		} else {
			t.Label = strings.TrimSpace(textExcludingLists(li))
		}
		for _, sub := range childLists(li) {
			t.Children = append(t.Children, topicsFrom(sub)...)
		}
		if t.Label != "" || t.Href != "" {
			out = append(out, t)
		}
	}
	return out
}

// childLists finds this item's sub-lists. They are not necessarily direct
// children: Asciidoctor wraps each list in <div class="ulist">, so the <ul>
// sits a level or two down. Descent stops at the first list found on a
// branch, leaving deeper nesting to the recursive call.
func childLists(li *html.Node) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			if c.Type != html.ElementNode {
				continue
			}
			if c.Data == "ul" || c.Data == "ol" {
				out = append(out, c)
				continue
			}
			walk(c)
		}
	}
	walk(li)
	return out
}

func firstLinkOutsideLists(n *html.Node) *html.Node {
	var found *html.Node
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		if found != nil {
			return
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				if c.Data == "ul" || c.Data == "ol" {
					continue
				}
				if c.Data == "a" {
					found = c
					return
				}
			}
			walk(c)
		}
	}
	walk(n)
	return found
}

func textExcludingLists(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && (c.Data == "ul" || c.Data == "ol") {
				continue
			}
			if c.Type == html.TextNode {
				b.WriteString(c.Data)
			}
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

func hrefOf(a *html.Node) string {
	for _, at := range a.Attr {
		if at.Key == "href" {
			return strings.TrimSpace(at.Val)
		}
	}
	return ""
}

func titleOf(doc *html.Node) string {
	var title string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if title != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return title
}

// navMeta reads <meta name="tobar-segais.*"> values and the document title
// from a generated navigation document. Asciidoctor emits these from a
// docinfo file with attribute substitution, so slug and version are declared
// once in the AsciiDoc source and carried into the bundle automatically.
func navMeta(raw []byte) (map[string]string, string) {
	doc, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return nil, ""
	}
	out := map[string]string{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, a := range n.Attr {
				switch a.Key {
				case "name":
					name = a.Val
				case "content":
					content = a.Val
				}
			}
			if strings.HasPrefix(name, metaPrefix) {
				if v := strings.TrimSpace(content); !unresolved(v) {
					out[strings.TrimPrefix(name, metaPrefix)] = v
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return out, strings.TrimSpace(titleOf(doc))
}

// atoi reads a whole number and gives 0 for anything else, which is the same
// answer as saying nothing: a priority that does not parse orders the bundle
// with everything that named no priority at all.
func atoi(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}

// unresolved reports a leftover attribute reference such as "{copyright}".
// Asciidoctor leaves these in place when the attribute is not set, which is
// what makes an optional line in a shared docinfo file possible: the meta is
// written unconditionally, and means nothing until the document defines the
// attribute.
func unresolved(v string) bool {
	return strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") &&
		!strings.ContainsAny(v[1:len(v)-1], "{} ")
}

// navIdentity reads whatever identity a generated navigation document
// declares. It may declare none, some or all of it; the caller fills the
// gaps from elsewhere.
func navIdentity(a *archive.Archive) (Identity, bool) {
	for _, name := range NavFiles {
		raw, err := a.ReadFile(name)
		if err != nil {
			continue
		}
		meta, title := navMeta(raw)
		slug := normalise(meta["slug"])
		// An explicit title beats the document's own, because some
		// converters invent one: Pandoc falls back to the filename when no
		// title metadata is given.
		if t := meta["title"]; t != "" {
			title = t
		}
		id := Identity{
			Slug:      slug,
			Version:   meta["version"],
			Title:     title,
			Hidden:    strings.EqualFold(meta["hidden"], "true"),
			Copyright: meta["copyright"],
			Priority:  atoi(meta["priority"]),
			Source:    "nav",
		}
		if al := meta["aliases"]; al != "" {
			for _, x := range strings.Split(al, ",") {
				if x = normalise(x); x != "" {
					id.Aliases = append(id.Aliases, x)
				}
			}
		}
		if id.Slug == "" && id.Version == "" && id.Title == "" && id.Copyright == "" &&
			id.Priority == 0 && len(id.Aliases) == 0 {
			continue // a navigation document that says nothing about identity
		}
		return id, true
	}
	return Identity{}, false
}
