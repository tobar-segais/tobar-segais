package search

import "strings"
import "testing"

func build() *Index {
	b := NewBuilder()
	b.Add(Doc{Slug: "m", Version: "1.16", Href: "intro.html", Title: "Introduction"},
		"Tobar Segais is a web container for Eclipse InfoCenter help bundles.")
	b.Add(Doc{Slug: "m", Version: "1.16", Href: "custom.html", Title: "Customization"},
		"You can customize the stylesheet and the navbar icon of the container.")
	b.Add(Doc{Slug: "m", Version: "1.16", Href: "build.html", Title: "Building"},
		"Building requires Maven and a Java compiler.")
	return b.Freeze()
}

func TestSearchRanksAndRequiresAllTerms(t *testing.T) {
	i := build()
	if got := i.Search("container", 10); len(got) != 2 {
		t.Fatalf("expected 2 hits for one term, got %d", len(got))
	}
	// Both terms must appear: only the customization page has both.
	got := i.Search("container stylesheet", 10)
	if len(got) != 1 || got[0].Href != "custom.html" {
		t.Fatalf("conjunctive search wrong: %+v", got)
	}
	// A term nobody has yields nothing, rather than everything.
	if got := i.Search("container zzzz", 10); len(got) != 0 {
		t.Fatalf("expected no hits, got %d", len(got))
	}
}

func TestTitleIsWeighted(t *testing.T) {
	i := build()
	got := i.Search("building", 10)
	if len(got) == 0 || got[0].Href != "build.html" {
		t.Fatalf("title match should rank first: %+v", got)
	}
}

func TestSnippetHighlights(t *testing.T) {
	i := build()
	got := i.Search("stylesheet", 1)
	if len(got) == 0 {
		t.Fatal("no hit")
	}
	if !strings.Contains(got[0].Snippet, MarkStart+"stylesheet"+MarkEnd) {
		t.Fatalf("match not marked: %q", got[0].Snippet)
	}
}
