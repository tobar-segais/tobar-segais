package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/tobar-segais/tobar-segais/internal/bundle"
	"github.com/tobar-segais/tobar-segais/internal/search"
	"github.com/tobar-segais/tobar-segais/internal/server"
)

func buildCmd() *cobra.Command {
	var (
		content   string
		out       string
		site      string
		deflt     string
		copyright string
		debug     bool
	)
	c := &cobra.Command{
		Use:     "build",
		Aliases: []string{"publish"},
		Short:   "Generate the whole site as static files",
		Long: `Generate the whole site as static files.

Every page is rendered with the same templates the server uses, so what is
generated is what would have been served. The output is plain files with no
server behind them, suitable for GitHub Pages or any static host.

  tobar-segais build --content ./content --out ./site

Search still works: the index is written alongside the pages and queried in
the browser.

Where the site is not hosted at the root of a domain -- a GitHub Pages project
site lives at /<repo>/ -- give the path it is rooted at:

  tobar-segais build --content ./content --out ./site --site /my-docs`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			level := slog.LevelInfo
			if debug {
				level = slog.LevelDebug
			}
			log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

			if _, err := os.Stat(content); err != nil {
				return err
			}

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			lib := bundle.NewLibrary(content, log)
			if err := lib.Reload(); err != nil {
				return err
			}

			// Generating is the one case where the index must be finished
			// before anything is written, since there is no server left
			// afterwards to catch up.
			ix := search.NewIndexer(log)
			go ix.Run(ctx)
			var all []*bundle.Bundle
			for _, p := range lib.Catalogue().Products {
				all = append(all, p.Versions...)
			}
			ix.Rebuild(all)
			for !ix.Status().Ready {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(10 * time.Millisecond):
				}
			}

			return server.GenerateStatic(lib, ix, server.StaticOptions{
				Out:         out,
				Site:        site,
				DefaultSlug: deflt,
				Copyright:   copyright,
			}, log)
		},
	}
	f := c.Flags()
	f.StringVar(&content, "content", "./content", "directory of .zip/.jar documentation archives")
	f.StringVarP(&out, "out", "o", "./site", "directory to write the generated site into")
	f.StringVar(&site, "site", "", "path the site is rooted at, if not the root of the domain")
	f.StringVar(&deflt, "default", "", "slug to serve at / instead of the catalogue index")
	f.StringVar(&copyright, "copyright", "", "notice about the documentation itself, shown in the footer")
	f.BoolVar(&debug, "debug", false, "verbose logging")
	return c
}
