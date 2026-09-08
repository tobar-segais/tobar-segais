package bundle

import "testing"
import "github.com/tobar-segais/tobar-segais/internal/archive"

func TestLoadTOCFromLegacyPlugin(t *testing.T) {
	a, err := archive.Open("../../demo/content/manual-1.16.jar")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	toc, err := LoadTOC(a)
	if err != nil {
		t.Fatal(err)
	}
	if toc.Label == "" {
		t.Error("no label")
	}
	var n int
	toc.Walk(func(*Topic) { n++ })
	t.Logf("label=%q topics=%d", toc.Label, n)
	if n == 0 {
		t.Fatal("no topics parsed")
	}
}

// A table of contents can name a landing page that is not one of its
// children. Anything that enumerates only the children misses it, which is
// how a generated site ended up redirecting to a page it had not written.
func TestPagesIncludesTheRootTopic(t *testing.T) {
	toc := &TOC{
		Label: "The Book",
		Href:  "book.html",
		Topics: []*Topic{
			{Label: "Intro", Href: "intro.html"},
			{Label: "Deep", Href: "guides/deep.html#section"},
			{Label: "Dup", Href: "intro.html"},
			{Label: "Group"},
		},
	}
	got := toc.Pages()
	want := []string{"book.html", "intro.html", "guides/deep.html"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// The real Eclipse bundle is exactly this shape.
func TestPagesOnRealBundle(t *testing.T) {
	a, err := archive.Open("../../demo/content/manual-1.16.jar")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	toc, err := LoadTOC(a)
	if err != nil {
		t.Fatal(err)
	}
	pages := toc.Pages()
	if len(pages) == 0 || pages[0] != "book-manual.html" {
		t.Fatalf("root page missing or not first: %v", pages)
	}
	for _, p := range pages {
		if _, ok := a.Lookup(p); !ok {
			t.Errorf("%s is listed but not in the archive", p)
		}
	}
	t.Logf("%d pages, first is %s", len(pages), pages[0])
}
