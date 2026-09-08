package server

import (
	"regexp"
	"strings"
	"testing"

	"github.com/tobar-segais/tobar-segais/internal/bundle"
)

// headingTags lists the heading levels in document order. Comparing these
// rather than exact markup keeps the tests robust against the ids the
// renderer generates.
func headingTags(html string) []string {
	return regexp.MustCompile(`<(h[1-6])[ >]`).FindAllString(html, -1)
}

// Content comes from archives the server did not necessarily build, so
// anything that would escape the page has to be removed at render time.
func TestBodyOfStripsWhatWouldEscapeThePage(t *testing.T) {
	in := []byte(`<!DOCTYPE html><html><head><title>A page</title>
		<style>body{background:red}</style></head>
		<body>
		  <style>.doc{display:none}</style>
		  <link rel="stylesheet" href="theme.css">
		  <base href="https://elsewhere.example/">
		  <script>alert(1)</script>
		  <h1 onclick="alert(2)">Heading</h1>
		  <p>Kept text.</p>
		  <noscript>hidden</noscript>
		</body></html>`)

	body, title, _ := bodyOf(in)
	got := string(body)

	if title != "A page" {
		t.Errorf("title = %q", title)
	}
	for _, gone := range []string{"<style", "stylesheet", "<base", "<script", "onclick", "<noscript", "display:none"} {
		if strings.Contains(strings.ToLower(got), strings.ToLower(gone)) {
			t.Errorf("%q survived:\n%s", gone, got)
		}
	}
	for _, kept := range []string{"Heading", "Kept text."} {
		if !strings.Contains(got, kept) {
			t.Errorf("%q was lost:\n%s", kept, got)
		}
	}
}

// Whatever a generator does with heading levels, a page reads as h1 then h2.
func TestBodyOfNormalisesHeadings(t *testing.T) {
	// Typst's shape: content starts at h2 because h1 is reserved.
	body, _, heads := bodyOf([]byte(`<html><body>
		<h2>Page title</h2><h3>First section</h3><h3>Second section</h3>
	</body></html>`))
	got := string(body)
	if tags := headingTags(got); strings.Join(tags, ",") != "<h1 ,<h2 ,<h2 " {
		t.Errorf("levels not normalised, got %v:\n%s", tags, got)
	}
	if !strings.Contains(got, ">Page title<") || !strings.Contains(got, ">First section<") {
		t.Errorf("heading text lost:\n%s", got)
	}
	// The page title is not listed in its own contents column.
	if len(heads) != 2 || heads[0].Text != "First section" {
		t.Fatalf("contents column wrong: %+v", heads)
	}
	if heads[0].ID == "" {
		t.Error("heading ids should be generated when absent")
	}
}

// A page that already starts at h1 is left alone.
func TestBodyOfLeavesCorrectHeadingsAlone(t *testing.T) {
	body, _, heads := bodyOf([]byte(`<html><body><h1>T</h1><h2>S</h2></body></html>`))
	if tags := headingTags(string(body)); strings.Join(tags, ",") != "<h1 ,<h2 " {
		t.Errorf("levels changed, got %v: %s", tags, body)
	}
	if len(heads) != 1 || heads[0].Text != "S" {
		t.Fatalf("contents column wrong: %+v", heads)
	}
}

// A page's own footer is chrome; a footer inside content is content. The
// header is never removed, because that is where generators put the title.
func TestStripChromeIsPositional(t *testing.T) {
	body, _, _ := bodyOf([]byte(`<html><body>
		<div id="header"><h1>Kept title</h1></div>
		<div id="content">
		  <article>
		    <p>Body text.</p>
		    <footer>Nested footer, part of the content.</footer>
		  </article>
		</div>
		<footer>Page footer, generator chrome.</footer>
		<div id="footer"><div id="footer-text">Last updated 2026</div></div>
	</body></html>`))
	got := string(body)

	if !strings.Contains(got, "Kept title") {
		t.Errorf("header removed, title lost:\n%s", got)
	}
	if !strings.Contains(got, "Nested footer") {
		t.Errorf("a footer inside content should survive:\n%s", got)
	}
	if strings.Contains(got, "Page footer") {
		t.Errorf("top-level footer should be removed:\n%s", got)
	}
	if strings.Contains(got, "Last updated") {
		t.Errorf("generator footer should be removed:\n%s", got)
	}
}

// The sidebar marks a nearest page only when there is one. A page missing
// from the root of a manual has no nearest sibling, and marking an arbitrary
// entry current would misreport where the reader is.
func TestNearestRequiresASharedPrefix(t *testing.T) {
	toc := &bundle.TOC{Topics: []*bundle.Topic{
		{Label: "Copyright", Href: "copyright.html"},
		{Label: "Guides", Children: []*bundle.Topic{
			{Label: "Installing", Href: "guides/install.html"},
		}},
	}}
	if got := nearest(toc, "formats.html"); got != "" {
		t.Errorf("root-level miss should mark nothing, got %q", got)
	}
	if got := nearest(toc, "guides/upgrade.html"); got != "guides/install.html" {
		t.Errorf("miss under guides/ should mark its sibling, got %q", got)
	}
}
