package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/tobar-segais/tobar-segais/internal/bundle"
	"github.com/tobar-segais/tobar-segais/internal/search"
	"github.com/tobar-segais/tobar-segais/internal/server"
)

func serveCmd() *cobra.Command {
	var (
		content   string
		addr      string
		poll      time.Duration
		deflt     string
		copyright string
		debug     bool
	)
	c := &cobra.Command{
		Use:   "serve",
		Short: "Serve documentation archives over HTTP",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			level := slog.LevelInfo
			if debug {
				level = slog.LevelDebug
			}
			log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

			if _, err := os.Stat(content); err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			lib := bundle.NewLibrary(content, log)
			ix := search.NewIndexer(log)

			// Indexing follows the catalogue rather than gating it: content
			// is servable as soon as the archives are open, and search
			// catches up behind.
			lib.OnChange = func(_ []*bundle.Bundle, cat *bundle.Catalogue) {
				var all []*bundle.Bundle
				for _, p := range cat.Products {
					all = append(all, p.Versions...)
				}
				ix.Rebuild(all)
			}

			go ix.Run(ctx)
			go bundle.Watch(ctx, lib, poll, log)

			srv, err := server.New(lib, ix, server.Options{DefaultSlug: deflt, Copyright: copyright}, log)
			if err != nil {
				return err
			}
			hs := &http.Server{
				Addr:              addr,
				Handler:           srv.Handler(),
				ReadHeaderTimeout: 10 * time.Second,
			}
			go func() {
				<-ctx.Done()
				sh, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				hs.Shutdown(sh)
			}()

			log.Info("listening", "addr", addr, "content", content)
			if err := hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	}
	f := c.Flags()
	f.StringVar(&content, "content", "./content", "directory of .zip/.jar documentation archives")
	f.StringVar(&addr, "addr", ":8080", "address to listen on")
	f.DurationVar(&poll, "poll", 30*time.Second, "content directory rescan interval, and the fallback when filesystem notifications are unavailable")
	f.StringVar(&deflt, "default", "", "slug to serve at / instead of the catalogue index")
	f.StringVar(&copyright, "copyright", "", "notice about the documentation itself, shown in the footer")
	f.BoolVar(&debug, "debug", false, "verbose logging")
	return c
}
