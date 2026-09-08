package bundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tobar-segais/tobar-segais/internal/archive"
)

// The archive names itself, badly, and cannot be rebuilt. The metadata file
// beside it is the fix, and it has to beat what the archive says rather than
// merely fill in around it.
func TestSidecarOverridesTheArchive(t *testing.T) {
	nav := `<!DOCTYPE html><html><head>
<title>Wrong Title</title>
<meta name="tobar-segais.slug" content="wrong-slug">
<meta name="tobar-segais.version" content="0.0.1">
<meta name="tobar-segais.copyright" content="Wrong notice">
</head><body><ul><li><a href="index.html">Introduction</a></li></ul></body></html>`

	dir := t.TempDir()
	p := filepath.Join(dir, "vendor-bundle.jar")
	if err := os.WriteFile(p, zipWith(t, map[string]string{"nav.html": nav}), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(SidecarPath(p), []byte(`
slug = "handbook"
version = "2.1.0"
title = "The Handbook"
aliases = ["old-handbook"]
copyright = "Copyright © 2026 Example Ltd."
`), 0o644); err != nil {
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
	if id.Slug != "handbook" || id.Version != "2.1.0" {
		t.Errorf("identity %q %q", id.Slug, id.Version)
	}
	if id.Title != "The Handbook" {
		t.Errorf("title %q", id.Title)
	}
	if id.Copyright != "Copyright © 2026 Example Ltd." {
		t.Errorf("copyright %q", id.Copyright)
	}
	if len(id.Aliases) != 1 || id.Aliases[0] != "old-handbook" {
		t.Errorf("aliases %v", id.Aliases)
	}
	if !strings.HasPrefix(id.Source, "sidecar") {
		t.Errorf("source %q should name the metadata file first", id.Source)
	}
}

// A metadata file that states one thing says nothing about the rest, so the
// archive still answers for everything it did not mention.
func TestSidecarFillsOnlyWhatItStates(t *testing.T) {
	nav := `<!DOCTYPE html><html><head>
<title>The Handbook</title>
<meta name="tobar-segais.slug" content="handbook">
<meta name="tobar-segais.version" content="1.0.0">
</head><body><ul><li><a href="index.html">Introduction</a></li></ul></body></html>`

	dir := t.TempDir()
	p := filepath.Join(dir, "handbook.zip")
	if err := os.WriteFile(p, zipWith(t, map[string]string{"nav.html": nav}), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(SidecarPath(p),
		[]byte("copyright = \"Copyright © 2026 Example Ltd.\"\n"), 0o644); err != nil {
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
	if id.Slug != "handbook" || id.Version != "1.0.0" || id.Title != "The Handbook" {
		t.Errorf("archive identity lost: %+v", id)
	}
	if id.Copyright != "Copyright © 2026 Example Ltd." {
		t.Errorf("copyright %q", id.Copyright)
	}
}

// hidden = false has to mean something, or a bundle that hid itself could
// never be shown.
func TestSidecarCanUnhide(t *testing.T) {
	nav := `<!DOCTYPE html><html><head>
<title>The Handbook</title>
<meta name="tobar-segais.slug" content="handbook">
<meta name="tobar-segais.version" content="1.0.0">
<meta name="tobar-segais.hidden" content="true">
</head><body><ul><li><a href="index.html">Introduction</a></li></ul></body></html>`

	dir := t.TempDir()
	p := filepath.Join(dir, "handbook.zip")
	if err := os.WriteFile(p, zipWith(t, map[string]string{"nav.html": nav}), 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := archive.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	if id, err := Identify(a, p); err != nil || !id.Hidden {
		t.Fatalf("expected the archive to hide itself: %+v %v", id, err)
	}
	if err := os.WriteFile(SidecarPath(p), []byte("hidden = false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	id, err := Identify(a, p)
	if err != nil {
		t.Fatal(err)
	}
	if id.Hidden {
		t.Error("metadata file could not unhide the bundle")
	}
}

// Someone wrote the file meaning to change something. Serving the archive's
// own identity instead would hide that it did nothing at all.
func TestMalformedSidecarIsAnError(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "handbook-1.0.0.zip")
	if err := os.WriteFile(p, zipWith(t, map[string]string{
		"nav.html": `<html><head><title>H</title></head><body><ul><li><a href="index.html">I</a></li></ul></body></html>`,
	}), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(SidecarPath(p), []byte("slug = handbook\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := archive.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	if _, err := Identify(a, p); err == nil {
		t.Error("expected a malformed metadata file to be reported")
	}
}

// Editing the metadata file must reload the archive beside it, even though
// the archive itself has not been touched.
func TestEditingSidecarReloads(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "handbook-1.0.0.zip")
	if err := os.WriteFile(p, zipWith(t, map[string]string{
		"nav.html": `<html><head><title>H</title></head><body><ul><li><a href="index.html">I</a></li></ul></body></html>`,
	}), 0o644); err != nil {
		t.Fatal(err)
	}

	l := NewLibrary(dir, quiet())
	t.Cleanup(l.Close) // Windows will not remove a file this still holds open
	if err := l.Reload(); err != nil {
		t.Fatal(err)
	}
	if _, ok := l.Catalogue().Product("handbook"); !ok {
		t.Fatal("expected handbook from the filename")
	}

	if err := os.WriteFile(SidecarPath(p), []byte("slug = \"manual\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := l.Reload(); err != nil {
		t.Fatal(err)
	}
	if _, ok := l.Catalogue().Product("manual"); !ok {
		t.Error("adding a metadata file did not reload the archive")
	}
	if _, ok := l.Catalogue().Product("handbook"); ok {
		t.Error("the old identity is still in the catalogue")
	}
}
