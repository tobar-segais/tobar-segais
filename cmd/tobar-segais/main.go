// Command tobar-segais serves and packages versioned documentation archives.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := root().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
