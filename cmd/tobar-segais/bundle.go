package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tobar-segais/tobar-segais/internal/archive"
	"github.com/tobar-segais/tobar-segais/internal/bundle"
	"github.com/tobar-segais/tobar-segais/internal/pack"
)

func bundleCmd() *cobra.Command {
	var (
		out     string
		format  string
		slug    string
		version string
		tool    string
		quiet   bool
	)
	c := &cobra.Command{
		Use:   "bundle <source-dir>",
		Short: "Package a documentation source directory into a bundle",
		Long: `Package a documentation source directory into a bundle.

The source language is detected from the files present, and converted with the
flags a bundle wants: content fragments for pages, and a navigation document
whose head carries the slug and version.

  asciidoc   nav.adoc  + *.adoc   (asciidoctor, or asciidoctor.js)
  markdown   nav.md    + *.md     (built in, nothing to install)
  typst      nav.typ   + *.typ    (typst)
  html       nav.html  + *.html   (copied as-is)

Where a converter is not on PATH under its usual name, point at it:

  tobar-segais bundle ./docs --tool "bundle exec asciidoctor"

The archive is named from the identity the bundle declares, so publishing is:

  tobar-segais bundle ./docs --out /srv/docs`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose := func(f string, a ...any) {
				if !quiet {
					fmt.Fprintf(cmd.OutOrStdout(), "  "+f+"\n", a...)
				}
			}
			tmp, err := pack.Build(pack.Options{
				Source:  args[0],
				OutDir:  out,
				Format:  pack.Format(format),
				Tool:    tool,
				Verbose: verbose,
			})
			if err != nil {
				return err
			}
			defer os.Remove(tmp)

			// Read the identity back out of the archive we just wrote, so
			// naming uses exactly the rules the server will apply to it.
			a, err := archive.Open(tmp)
			if err != nil {
				return err
			}
			id, err := bundle.Identify(a, tmp)
			a.Close()
			if err != nil {
				return fmt.Errorf("cannot identify the bundle: %w", err)
			}
			if slug != "" {
				id.Slug = slug
			}
			if version != "" {
				id.Version = version
			}
			if id.Slug == "" || id.Version == "" {
				return fmt.Errorf("bundle has no %s; declare them in the navigation document or pass --slug and --version",
					map[bool]string{true: "version", false: "slug"}[id.Slug != ""])
			}

			final := filepath.Join(out, id.Slug+"-"+id.Version+".zip")
			if err := os.Rename(tmp, final); err != nil {
				return err
			}
			st, _ := os.Stat(final)
			fmt.Fprintf(cmd.OutOrStdout(), "%s (%s %s, from %s, %d bytes)\n",
				final, id.Slug, id.Version, id.Source, st.Size())
			return nil
		},
	}
	f := c.Flags()
	f.StringVarP(&out, "out", "o", "./content", "directory to write the bundle into")
	f.StringVar(&format, "format", "", "source format: asciidoc, markdown, typst or html (default: detect)")
	f.StringVar(&tool, "tool", "", "command to run the converter, if it is not on PATH under its usual name; may include arguments, e.g. \"bundle exec asciidoctor\"")
	f.StringVar(&slug, "slug", "", "override the slug declared by the source")
	f.StringVar(&version, "version", "", "override the version declared by the source")
	f.BoolVarP(&quiet, "quiet", "q", false, "only print the resulting path")
	return c
}
