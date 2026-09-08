package bundle

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
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
