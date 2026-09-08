package bundle

import (
	"os"
	"testing"

	"github.com/tobar-segais/tobar-segais/internal/archive"
)

// The exact shape Asciidoctor emits: item text wrapped in <p>, sub-lists
// nested inside the parent <li>, and a grouping item with no link.
const asciidoctorNav = `<!DOCTYPE html><html><head>
<title>Tobar Segais: User Manual</title>
<meta name="tobar-segais.slug" content="tobar-segais">
<meta name="tobar-segais.version" content="2.0.0">
</head><body><div id="content"><div class="ulist"><ul>
<li><p><a href="index.html">Introduction</a></p></li>
<li><p>Publishing documentation</p><div class="ulist"><ul>
  <li><p><a href="bundles.html">Bundle format</a></p></li>
  <li><p><a href="asciidoctor.html">Writing with Asciidoctor</a></p></li>
</ul></div></li>
<li><p><a href="legacy.html">Legacy Eclipse bundles</a></p></li>
</ul></div></div></body></html>`

func TestParseNavHTML(t *testing.T) {
	toc, err := ParseNavHTML([]byte(asciidoctorNav))
	if err != nil {
		t.Fatal(err)
	}
	if toc.Label != "Tobar Segais: User Manual" {
		t.Errorf("label %q", toc.Label)
	}
	if len(toc.Topics) != 3 {
		t.Fatalf("expected 3 top-level topics, got %d", len(toc.Topics))
	}
	if toc.Topics[0].Href != "index.html" || toc.Topics[0].Label != "Introduction" {
		t.Errorf("first topic wrong: %+v", toc.Topics[0])
	}
	// A grouping item has a label but no link, and carries children.
	g := toc.Topics[1]
	if g.Href != "" || g.Label != "Publishing documentation" {
		t.Errorf("grouping item wrong: %+v", g)
	}
	if len(g.Children) != 2 || g.Children[1].Href != "asciidoctor.html" {
		t.Errorf("nested children wrong: %+v", g.Children)
	}
	// The parent's label must not absorb its children's text.
	if g.Label == "" || len(g.Label) > 40 {
		t.Errorf("parent label picked up child text: %q", g.Label)
	}
}

func TestNavHTMLProvidesIdentity(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/anything.zip"
	if err := os.WriteFile(p, zipWith(t, map[string]string{"nav.html": asciidoctorNav}), 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := archive.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	id, err := Identify(a, p)
	if err != nil {
		t.Fatal(err)
	}
	if id.Source != "nav" || id.Slug != "tobar-segais" || id.Version != "2.0.0" {
		t.Fatalf("identity %+v", id)
	}
	if id.Title != "Tobar Segais: User Manual" {
		t.Errorf("title %q", id.Title)
	}
	// And the same file supplies the table of contents.
	toc, err := LoadTOC(a)
	if err != nil {
		t.Fatal(err)
	}
	if len(toc.Topics) != 3 {
		t.Errorf("toc not loaded from nav.html: %+v", toc)
	}
}

func TestNavHTMLCarriesCopyright(t *testing.T) {
	nav := `<!DOCTYPE html><html><head>
<title>The Handbook</title>
<meta name="tobar-segais.slug" content="handbook">
<meta name="tobar-segais.version" content="1.0.0">
<meta name="tobar-segais.copyright" content="Copyright © 2026 Example Ltd.">
</head><body><ul><li><a href="index.html">Introduction</a></li></ul></body></html>`

	dir := t.TempDir()
	p := dir + "/handbook.zip"
	if err := os.WriteFile(p, zipWith(t, map[string]string{"nav.html": nav}), 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := archive.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	id, err := Identify(a, p)
	if err != nil {
		t.Fatal(err)
	}
	if id.Copyright != "Copyright © 2026 Example Ltd." {
		t.Errorf("copyright %q", id.Copyright)
	}
}

// An optional line in a shared docinfo file is written unconditionally, so
// Asciidoctor leaves "{copyright}" in the output when the document does not
// define the attribute. That is not a notice, and must not be treated as one.
func TestUnresolvedAttributeIsNotMetadata(t *testing.T) {
	nav := `<!DOCTYPE html><html><head>
<title>The Handbook</title>
<meta name="tobar-segais.slug" content="handbook">
<meta name="tobar-segais.copyright" content="{copyright}">
</head><body><ul><li><a href="index.html">Introduction</a></li></ul></body></html>`

	dir := t.TempDir()
	p := dir + "/handbook-1.0.0.zip"
	if err := os.WriteFile(p, zipWith(t, map[string]string{"nav.html": nav}), 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := archive.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	id, err := Identify(a, p)
	if err != nil {
		t.Fatal(err)
	}
	if id.Copyright != "" {
		t.Errorf("kept an unresolved attribute reference: %q", id.Copyright)
	}
	if id.Slug != "handbook" {
		t.Errorf("slug %q", id.Slug)
	}
}
