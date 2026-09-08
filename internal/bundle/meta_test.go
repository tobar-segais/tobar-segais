package bundle

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/tobar-segais/tobar-segais/internal/archive"
)

// makeZip builds an archive on disk from name->content pairs.
func makeZip(t *testing.T, name string, files map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for n, c := range files {
		w, err := zw.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(c)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return p
}

func identify(t *testing.T, p string) Identity {
	t.Helper()
	a, err := archive.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	id, err := Identify(a, p)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// With the metadata file gone, OSGi headers are the first fallback after the
// navigation document.
func TestIdentifyFallsBackToOSGi(t *testing.T) {
	p := makeZip(t, "whatever.jar", map[string]string{
		"META-INF/MANIFEST.MF":              "Bundle-SymbolicName: com.example.help;singleton:=true\nBundle-Version: 3.4.5\n",
		"META-INF/maven/x/y/pom.properties": "artifactId=ignored\nversion=0.1\n",
	})
	id := identify(t, p)
	if id.Source != "osgi" || id.Slug != "com-example-help" || id.Version != "3.4.5" {
		t.Fatalf("got %+v", id)
	}
}

func TestIdentifyFallsBackToMaven(t *testing.T) {
	// The real bundle from Maven Central: no metadata file, no OSGi headers.
	id := identify(t, "../../demo/content/manual-1.16.jar")
	if id.Source != "maven" || id.Slug != "tobar-segais-manual" || id.Version != "1.16" {
		t.Fatalf("got %+v", id)
	}
	if v, ok := id.Semver(); !ok || v.String() != "1.16.0" {
		t.Errorf("lenient semver failed: %v %v", v, ok)
	}
}

func TestIdentifyFallsBackToFilename(t *testing.T) {
	p := makeZip(t, "acme-docs-4.2.1.zip", map[string]string{"index.html": "hi"})
	id := identify(t, p)
	if id.Source != "filename" || id.Slug != "acme-docs" || id.Version != "4.2.1" {
		t.Fatalf("got %+v", id)
	}
}

func TestSemverPrecedence(t *testing.T) {
	// Semver rules, not OSGi: a pre-release sorts before its release, so a
	// SNAPSHOT never outranks the version it precedes.
	older := Identity{Version: "1.16.0-SNAPSHOT"}
	newer := Identity{Version: "1.16.0"}
	a, _ := older.Semver()
	b, _ := newer.Semver()
	if !a.LessThan(b) {
		t.Fatalf("%s should sort before %s", a, b)
	}
}
