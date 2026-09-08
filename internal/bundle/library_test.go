package bundle

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/tobar-segais/tobar-segais/internal/archive"
)

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func copyTo(t *testing.T, src, dst string) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLibraryGroupsVersionsNewestFirst(t *testing.T) {
	dir := t.TempDir()
	copyTo(t, "../../demo/content/manual-1.15.jar", filepath.Join(dir, "manual-1.15.jar"))
	copyTo(t, "../../demo/content/manual-1.16.jar", filepath.Join(dir, "manual-1.16.jar"))

	l := NewLibrary(dir, quiet())
	t.Cleanup(l.Close) // Windows will not remove a file this still holds open
	if err := l.Reload(); err != nil {
		t.Fatal(err)
	}
	cat := l.Catalogue()
	if len(cat.Products) != 1 {
		t.Fatalf("expected one product, got %d", len(cat.Products))
	}
	p := cat.Products[0]
	if p.Slug != "tobar-segais-manual" {
		t.Fatalf("slug %q", p.Slug)
	}
	if len(p.Versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(p.Versions))
	}
	if p.Versions[0].Version != "1.16" || p.Versions[1].Version != "1.15" {
		t.Fatalf("wrong order: %s then %s", p.Versions[0].Version, p.Versions[1].Version)
	}
	if p.Latest().Version != "1.16" {
		t.Errorf("latest is %s", p.Latest().Version)
	}
	if p.Find("1.15") == nil {
		t.Error("1.15 not findable")
	}
}

// Adding and removing archives must take effect without a restart.
func TestWatchPicksUpAddAndRemove(t *testing.T) {
	dir := t.TempDir()
	l := NewLibrary(dir, quiet())
	t.Cleanup(l.Close) // Windows will not remove a file this still holds open
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go Watch(ctx, l, 100*time.Millisecond, quiet())

	added := filepath.Join(dir, "manual-1.16.jar")
	copyTo(t, "../../demo/content/manual-1.16.jar", added)

	if !eventually(t, 3*time.Second, func() bool {
		_, ok := l.Catalogue().Product("tobar-segais-manual")
		return ok
	}) {
		t.Fatal("added archive was never picked up")
	}

	if err := os.Remove(added); err != nil {
		t.Fatal(err)
	}
	if !eventually(t, 3*time.Second, func() bool {
		_, ok := l.Catalogue().Product("tobar-segais-manual")
		return !ok
	}) {
		t.Fatal("removed archive was never dropped")
	}
}

func eventually(t *testing.T, limit time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(limit)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

// The catalogue is ordered by priority, and alphabetically within a priority.
// A landing page is the reason this exists: it has to come first, and its
// title rarely puts it there.
func TestCatalogueOrdersByPriorityThenTitle(t *testing.T) {
	dir := t.TempDir()
	write := func(slug, title string, priority int) {
		nav := "<!DOCTYPE html><html><head><title>" + title + "</title>" +
			`<meta name="tobar-segais.slug" content="` + slug + `">` +
			`<meta name="tobar-segais.version" content="1.0.0">` +
			`<meta name="tobar-segais.priority" content="` + strconv.Itoa(priority) + `">` +
			`</head><body><ul><li><a href="index.html">Introduction</a></li></ul></body></html>`
		p := filepath.Join(dir, slug+"-1.0.0.zip")
		if err := os.WriteFile(p, zipWith(t, map[string]string{"nav.html": nav}), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Alphabetically this is apples, welcome, zebra. By priority it is not.
	write("zebra", "Zebra", 0)
	write("apples", "Apples", 0)
	write("welcome", "Welcome", 100)
	write("attic", "Attic", -10)

	l := NewLibrary(dir, quiet())
	t.Cleanup(l.Close)
	if err := l.Reload(); err != nil {
		t.Fatal(err)
	}

	var got []string
	for _, p := range l.Catalogue().Products {
		got = append(got, p.Slug)
	}
	want := []string{"welcome", "apples", "zebra", "attic"}
	if !slices.Equal(got, want) {
		t.Errorf("order is %v, want %v", got, want)
	}
}

// A metadata file beside an archive can order a bundle that says nothing about
// where it belongs -- and can move one that does.
func TestSidecarSetsPriority(t *testing.T) {
	dir := t.TempDir()
	nav := `<!DOCTYPE html><html><head><title>Handbook</title>` +
		`<meta name="tobar-segais.slug" content="handbook">` +
		`<meta name="tobar-segais.version" content="1.0.0">` +
		`<meta name="tobar-segais.priority" content="50">` +
		`</head><body><ul><li><a href="index.html">Introduction</a></li></ul></body></html>`
	p := filepath.Join(dir, "handbook-1.0.0.zip")
	if err := os.WriteFile(p, zipWith(t, map[string]string{"nav.html": nav}), 0o644); err != nil {
		t.Fatal(err)
	}

	a, err := archive.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if id, _ := Identify(a, p); id.Priority != 50 {
		t.Fatalf("priority from the bundle is %d", id.Priority)
	}

	// Zero has to mean zero here, or a bundle that raised itself could never
	// be put back among the rest.
	if err := os.WriteFile(SidecarPath(p), []byte("priority = 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	id, err := Identify(a, p)
	if err != nil {
		t.Fatal(err)
	}
	if id.Priority != 0 {
		t.Errorf("the metadata file did not put it back: priority %d", id.Priority)
	}
}
