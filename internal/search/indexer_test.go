package search

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tobar-segais/tobar-segais/internal/bundle"
)

// The real manual bundle must index and become searchable.
func TestIndexerIndexesRealBundle(t *testing.T) {
	dir := t.TempDir()
	raw, err := os.ReadFile("../../demo/content/manual-1.16.jar")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manual-1.16.jar"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	l := bundle.NewLibrary(dir, log)
	if err := l.Reload(); err != nil {
		t.Fatal(err)
	}

	ix := NewIndexer(log)
	// Before indexing, search is empty but the server is already usable.
	if ix.Index().Len() != 0 || ix.Status().Ready {
		t.Fatal("expected an empty, not-ready index at startup")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ix.Run(ctx)

	var all []*bundle.Bundle
	for _, p := range l.Catalogue().Products {
		all = append(all, p.Versions...)
	}
	ix.Rebuild(all)

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && !ix.Status().Ready {
		time.Sleep(10 * time.Millisecond)
	}
	st := ix.Status()
	if !st.Ready {
		t.Fatal("index never became ready")
	}
	if ix.Index().Len() == 0 {
		t.Fatal("no documents indexed")
	}
	t.Logf("indexed %d documents (%d/%d pages)", ix.Index().Len(), st.Done, st.Total)

	hits := ix.Index().Search("customization", 5)
	if len(hits) == 0 {
		t.Fatal("expected hits for a word that is in the manual")
	}
	t.Logf("top hit: %s (%s) %.2f", hits[0].Title, hits[0].Href, hits[0].Score)
}
