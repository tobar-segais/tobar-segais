// Package archive serves files straight out of a zip without unpacking it.
//
// Opening a zip parses its central directory, which already carries every
// entry's sizes, CRC-32 and method. What it does not give away is where an
// entry's data actually starts: that needs the local file header, which
// archive/zip re-reads on every Open because the header carries its own
// variable-length name and extra fields. We resolve that offset once per
// entry at load and keep it, so a request is a seek and a copy.
package archive

import (
	"archive/zip"
	"compress/flate"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

// Entry is everything needed to serve a file without touching the zip
// structure again.
type Entry struct {
	Name     string
	DataOff  int64  // offset of the compressed bytes in the file
	CompSize int64  // bytes on disk
	Size     int64  // bytes once inflated
	CRC32    uint32 // of the uncompressed data
	Method   uint16 // zip.Store or zip.Deflate
	Modified time.Time
}

// Stored reports whether the entry is uncompressed, and so directly seekable.
func (e *Entry) Stored() bool { return e.Method == zip.Store }

// ETag is derived from the CRC-32 the archive already stores, so conditional
// requests are exact rather than heuristic.
func (e *Entry) ETag() string { return fmt.Sprintf(`"%08x-%x"`, e.CRC32, e.Size) }

// Archive is an open zip with a resolved offset table.
type Archive struct {
	Path    string
	file    *os.File
	size    int64
	entries map[string]*Entry
	names   []string
}

// Open reads the central directory and resolves every entry's data offset.
// This is one pass at startup; nothing here scales with archive content size.
func Open(name string) (*Archive, error) {
	f, err := openShared(name)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	zr, err := zip.NewReader(f, st.Size())
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("%s: %w", name, err)
	}

	a := &Archive{
		Path:    name,
		file:    f,
		size:    st.Size(),
		entries: make(map[string]*Entry, len(zr.File)),
		names:   make([]string, 0, len(zr.File)),
	}
	for _, zf := range zr.File {
		if strings.HasSuffix(zf.Name, "/") {
			continue // directory marker
		}
		if zf.Method != zip.Store && zf.Method != zip.Deflate {
			// Nothing else can be served without a decompressor we do not
			// have; skip rather than fail the whole archive.
			continue
		}
		off, err := zf.DataOffset()
		if err != nil {
			continue // truncated or otherwise unreadable entry
		}
		clean := path.Clean("/" + zf.Name)[1:]
		e := &Entry{
			Name:     clean,
			DataOff:  off,
			CompSize: int64(zf.CompressedSize64),
			Size:     int64(zf.UncompressedSize64),
			CRC32:    zf.CRC32,
			Method:   zf.Method,
			Modified: zf.Modified,
		}
		a.entries[clean] = e
		a.names = append(a.names, clean)
	}
	sort.Strings(a.names)
	return a, nil
}

func (a *Archive) Close() error { return a.file.Close() }

// Lookup finds an entry by its cleaned path.
func (a *Archive) Lookup(name string) (*Entry, bool) {
	e, ok := a.entries[path.Clean("/" + name)[1:]]
	return e, ok
}

// Names lists every servable entry, sorted.
func (a *Archive) Names() []string { return a.names }

// Raw returns the entry's bytes exactly as they sit on disk: the deflate
// stream for compressed entries, the content itself for stored ones. It is a
// SectionReader over a shared *os.File, so concurrent callers are fine and
// no lock is needed.
func (a *Archive) Raw(e *Entry) *io.SectionReader {
	return io.NewSectionReader(a.file, e.DataOff, e.CompSize)
}

// gzipHeader is a fixed 10-byte header: magic, deflate, no flags, no mtime,
// no extra flags, unknown OS.
var gzipHeader = [10]byte{0x1f, 0x8b, 8, 0, 0, 0, 0, 0, 0, 0xff}

// ErrNotDeflated is returned when a gzip passthrough is requested for an
// entry that is not deflate-compressed.
var ErrNotDeflated = errors.New("archive: entry is not deflated")

// GzipPassthrough writes the entry as a gzip stream without decompressing
// it. A gzip member is a header, a raw deflate stream, then the CRC-32 and
// the uncompressed size -- and a zip entry is a raw deflate stream whose
// CRC-32 and size the central directory already recorded. So the body is a
// byte copy from disk to the wire: no inflate, no recompress.
func (a *Archive) GzipPassthrough(w io.Writer, e *Entry) error {
	if e.Method != zip.Deflate {
		return ErrNotDeflated
	}
	if _, err := w.Write(gzipHeader[:]); err != nil {
		return err
	}
	if _, err := io.Copy(w, a.Raw(e)); err != nil {
		return err
	}
	var trailer [8]byte
	binary.LittleEndian.PutUint32(trailer[0:4], e.CRC32)
	binary.LittleEndian.PutUint32(trailer[4:8], uint32(e.Size))
	_, err := w.Write(trailer[:])
	return err
}

// GzipSize is the exact number of bytes GzipPassthrough will write, so the
// response can carry a Content-Length.
func GzipSize(e *Entry) int64 { return int64(len(gzipHeader)) + e.CompSize + 8 }

// ReadFile inflates an entry into memory. Used by the indexer and for small
// metadata files, not on the serving path for large content.
func (a *Archive) ReadFile(name string) ([]byte, error) {
	e, ok := a.Lookup(name)
	if !ok {
		return nil, os.ErrNotExist
	}
	rc, err := a.OpenEntry(e)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	buf := make([]byte, e.Size)
	if _, err := io.ReadFull(rc, buf); err != nil {
		return nil, err
	}
	if crc32.ChecksumIEEE(buf) != e.CRC32 {
		return nil, fmt.Errorf("%s: %s: checksum mismatch", a.Path, name)
	}
	return buf, nil
}

// OpenEntry returns the entry's decompressed contents. Stored entries are
// served straight from the section reader; deflated ones are inflated.
func (a *Archive) OpenEntry(e *Entry) (io.ReadCloser, error) {
	sr := a.Raw(e)
	if e.Stored() {
		return io.NopCloser(sr), nil
	}
	return flate.NewReader(sr), nil
}
