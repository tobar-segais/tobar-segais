package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tobar-segais/tobar-segais/internal/version"
)

// The release workflow tests this output, so the shape of the line is part of
// the release: a rename here that silently stopped the -ldflags stamp landing
// would otherwise ship binaries that report a commit hash.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			fmt.Fprintf(c.OutOrStdout(), "tobar-segais %s\n", version.String())
			return nil
		},
	}
}
