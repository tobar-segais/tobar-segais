package bundle

import (
	"encoding/xml"
	"fmt"
	"path"
	"strings"

	"github.com/tobar-segais/tobar-segais/internal/archive"
)

// Topic is one node of a table of contents. Href is relative to the archive
// root and empty for a grouping node that is not itself a page.
type Topic struct {
	Label    string
	Href     string
	Children []*Topic
}

// TOC is a bundle's navigation tree.
type TOC struct {
	Label  string
	Href   string
	Topics []*Topic
}

// --- Eclipse toc.xml -------------------------------------------------------

type xmlTopic struct {
	Label  string     `xml:"label,attr"`
	Href   string     `xml:"href,attr"`
	Topics []xmlTopic `xml:"topic"`
}

type xmlTOC struct {
	Label  string     `xml:"label,attr"`
	Topic  string     `xml:"topic,attr"`
	Topics []xmlTopic `xml:"topic"`
}

func convert(in []xmlTopic) []*Topic {
	out := make([]*Topic, 0, len(in))
	for _, t := range in {
		out = append(out, &Topic{
			Label:    strings.TrimSpace(t.Label),
			Href:     strings.TrimSpace(t.Href),
			Children: convert(t.Topics),
		})
	}
	return out
}

// ParseTOC reads an Eclipse toc.xml.
func ParseTOC(raw []byte) (*TOC, error) {
	var x xmlTOC
	if err := xml.Unmarshal(raw, &x); err != nil {
		return nil, err
	}
	return &TOC{
		Label:  strings.TrimSpace(x.Label),
		Href:   strings.TrimSpace(x.Topic),
		Topics: convert(x.Topics),
	}, nil
}

// --- legacy plugin.xml -----------------------------------------------------

type xmlPlugin struct {
	Extensions []struct {
		Point string `xml:"point,attr"`
		TOC   []struct {
			File    string `xml:"file,attr"`
			Primary bool   `xml:"primary,attr"`
		} `xml:"toc"`
		Index []struct {
			File string `xml:"file,attr"`
		} `xml:"index"`
	} `xml:"extension"`
}

// tocFromPlugin finds the toc file declared by an Eclipse plugin.xml,
// preferring the one marked primary.
func tocFromPlugin(raw []byte) (toc, index string) {
	var p xmlPlugin
	if err := xml.Unmarshal(raw, &p); err != nil {
		return "", ""
	}
	for _, ext := range p.Extensions {
		switch ext.Point {
		case "org.eclipse.help.toc":
			for _, t := range ext.TOC {
				if t.Primary && t.File != "" {
					toc = t.File
				} else if toc == "" && t.File != "" {
					toc = t.File
				}
			}
		case "org.eclipse.help.index":
			for _, i := range ext.Index {
				if i.File != "" && index == "" {
					index = i.File
				}
			}
		}
	}
	return toc, index
}

// LoadTOC resolves and parses a bundle's table of contents.
//
// The convention is a nav.html at the archive root: a generated navigation
// document holding the tree, and the bundle's identity in its head. Failing
// that we read a legacy Eclipse plugin.xml and follow its
// org.eclipse.help.toc extension, and failing that a toc.xml.
func LoadTOC(a *archive.Archive) (*TOC, error) {
	candidates := append([]string{}, NavFiles...)
	if raw, err := a.ReadFile("plugin.xml"); err == nil {
		if toc, _ := tocFromPlugin(raw); toc != "" {
			candidates = append(candidates, toc)
		}
	}
	candidates = append(candidates, "toc.xml")

	for _, c := range candidates {
		raw, err := a.ReadFile(c)
		if err != nil {
			continue
		}
		toc, err := parseTOCFile(c, raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", c, err)
		}
		return toc, nil
	}
	return nil, fmt.Errorf("no navigation: provide a nav.html at the archive root, or a legacy Eclipse toc.xml")
}

// Pages lists every page the table of contents refers to, in reading order.
//
// The root topic comes first and is easy to miss: a table of contents can
// name a landing page that is not also one of its children -- Eclipse
// bundles routinely do, with <toc topic="book.html"> -- and anything that
// enumerates only the children will not know that page exists.
func (t *TOC) Pages() []string {
	var out []string
	seen := map[string]bool{}
	add := func(href string) {
		if i := strings.IndexAny(href, "#?"); i != -1 {
			href = href[:i]
		}
		if href == "" || seen[href] {
			return
		}
		seen[href] = true
		out = append(out, href)
	}
	add(t.Href)
	t.Walk(func(x *Topic) { add(x.Href) })
	return out
}

// Walk visits every topic depth first.
func (t *TOC) Walk(fn func(*Topic)) {
	var rec func([]*Topic)
	rec = func(ts []*Topic) {
		for _, x := range ts {
			fn(x)
			rec(x.Children)
		}
	}
	rec(t.Topics)
}

// parseTOCFile dispatches on extension: HTML is the convention, XML the
// legacy Eclipse format.
func parseTOCFile(name string, raw []byte) (*TOC, error) {
	switch strings.ToLower(path.Ext(name)) {
	case ".html", ".htm":
		return ParseNavHTML(raw)
	default:
		return ParseTOC(raw)
	}
}
