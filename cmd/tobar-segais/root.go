package main

import "github.com/spf13/cobra"

func root() *cobra.Command {
	c := &cobra.Command{
		Use:   "tobar-segais",
		Short: "Serve and package versioned documentation archives",
		Long: `Tobar Segais serves versioned documentation archives.

Point it at a directory of .zip or .jar bundles and it publishes each one at a
stable URL, with navigation, full-text search and a version picker.

  tobar-segais serve  --content ./content
  tobar-segais bundle ./docs --out ./content
  tobar-segais build  --content ./content --out ./site`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	c.AddCommand(serveCmd(), bundleCmd(), buildCmd(), newVersionCmd())
	return c
}
