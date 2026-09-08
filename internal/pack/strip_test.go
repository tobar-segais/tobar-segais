package pack

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripStylesRemovesCSS(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "page.html")
	in := `<!DOCTYPE html><html><head><title>T</title>` +
		`<link rel="stylesheet" href="x.css">` +
		`<style>body{color:red}</style></head>` +
		`<body><h1>Kept</h1><p>Also kept</p></body></html>`
	if err := os.WriteFile(p, []byte(in), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stripStyles(dir); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(p)
	for _, gone := range []string{"<style", "stylesheet", "color:red"} {
		if bytes.Contains(bytes.ToLower(out), []byte(gone)) {
			t.Errorf("%q survived: %s", gone, out)
		}
	}
	for _, kept := range []string{"<title>T</title>", "Kept", "Also kept"} {
		if !bytes.Contains(out, []byte(kept)) {
			t.Errorf("%q was lost: %s", kept, out)
		}
	}
}

// Files with no CSS must not be rewritten at all: re-rendering a fragment
// would wrap it in html/head/body for no reason.
func TestStripStylesLeavesCleanFilesByteIdentical(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "frag.html")
	in := "<h1>Title</h1>\n<p>A fragment, unwrapped.</p>\n"
	if err := os.WriteFile(p, []byte(in), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stripStyles(dir); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(p)
	if string(out) != in {
		t.Errorf("clean file was rewritten:\n got %q\nwant %q", out, in)
	}
}

// Non-HTML files are not touched.
func TestStripStylesIgnoresOtherFiles(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "theme.css")
	in := "body{color:red}"
	if err := os.WriteFile(p, []byte(in), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := stripStyles(dir); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(p)
	if string(out) != in {
		t.Errorf("css asset was modified: %q", out)
	}
}

// Files whose names begin with an underscore are partials: pulled into other
// pages, never pages themselves.
func TestSourceFilesSkipsPartials(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"index.typ", "_common.typ", "chapter.typ", "_preamble.typ"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("= x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := sourceFiles(dir, Typst)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, p := range got {
		names = append(names, filepath.Base(p))
	}
	want := "chapter.typ,index.typ"
	if strings.Join(names, ",") != want {
		t.Errorf("got %v, want %s", names, want)
	}
}
