package archive

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"
)

const testJar = "../../demo/content/manual-1.16.jar"

func TestOpenResolvesOffsets(t *testing.T) {
	a, err := Open(testJar)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if len(a.Names()) == 0 {
		t.Fatal("no entries")
	}
	for _, n := range a.Names() {
		e, ok := a.Lookup(n)
		if !ok {
			t.Fatalf("%s: listed but not found", n)
		}
		if e.DataOff <= 0 {
			t.Errorf("%s: data offset not resolved", n)
		}
	}
}

// The whole point of the passthrough: gzip-decoding what we emit must give
// back exactly what inflating the entry gives.
func TestGzipPassthroughRoundTrips(t *testing.T) {
	a, err := Open(testJar)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	var checked int
	for _, n := range a.Names() {
		e, _ := a.Lookup(n)
		if e.Stored() {
			continue
		}
		want, err := a.ReadFile(n) // inflates, and verifies CRC
		if err != nil {
			t.Fatalf("%s: %v", n, err)
		}

		var buf bytes.Buffer
		if err := a.GzipPassthrough(&buf, e); err != nil {
			t.Fatalf("%s: %v", n, err)
		}
		if int64(buf.Len()) != GzipSize(e) {
			t.Errorf("%s: wrote %d bytes, GzipSize said %d", n, buf.Len(), GzipSize(e))
		}

		zr, err := gzip.NewReader(&buf)
		if err != nil {
			t.Fatalf("%s: gzip header rejected: %v", n, err)
		}
		got, err := io.ReadAll(zr)
		if err != nil {
			t.Fatalf("%s: gzip body/trailer rejected: %v", n, err)
		}
		if err := zr.Close(); err != nil {
			t.Fatalf("%s: gzip close (checksum) failed: %v", n, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%s: round trip differs (%d vs %d bytes)", n, len(got), len(want))
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no deflated entries exercised")
	}
	t.Logf("round-tripped %d deflated entries", checked)
}
