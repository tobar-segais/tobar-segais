//go:build !windows

package archive

import "os"

// openShared opens a file that can still be deleted or replaced while it is
// open, which everywhere but Windows is what an ordinary open already gives.
func openShared(name string) (*os.File, error) { return os.Open(name) }
