package archive

import (
	"os"

	"golang.org/x/sys/windows"
)

// openShared opens a file that can still be deleted or replaced while it is
// open.
//
// The whole design keeps an archive open: the offsets inside it are resolved
// once, and every response is read from that handle. Windows refuses to unlink
// a file that a process holds open unless the handle allows it, and os.Open
// does not ask for FILE_SHARE_DELETE. Without this, dropping a new version of
// an archive into the watched directory -- the thing the watcher exists for --
// fails on Windows with "the process cannot access the file".
//
// Deleting a file that is still open leaves the reader with the bytes it
// already had, which is exactly what the server needs: it serves the old
// archive until the reload swaps in the new catalogue and closes the handle.
func openShared(name string) (*os.File, error) {
	p, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: name, Err: err}
	}
	h, err := windows.CreateFile(
		p,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: name, Err: err}
	}
	return os.NewFile(uintptr(h), name), nil
}
