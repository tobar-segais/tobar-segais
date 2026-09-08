// Package pack turns a directory of documentation source into a bundle.
//
// The knowledge that earns its keep here is the set of flags each converter
// needs to produce output a bundle wants: content fragments rather than whole
// documents, and a navigation document whose head carries the identity. Those
// flags are not obvious, and they should not have to be rediscovered in a
// shell script per project.
package pack

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Format is a documentation source language.
type Format string

const (
	AsciiDoc Format = "asciidoc"
	Markdown Format = "markdown"
	Typst    Format = "typst"
	HTML     Format = "html"
)

// Formats lists what can be packed, for help text.
var Formats = []Format{AsciiDoc, Markdown, Typst, HTML}

var extensions = map[Format][]string{
	AsciiDoc: {".adoc", ".asciidoc", ".adoc.txt"},
	Markdown: {".md", ".markdown"},
	Typst:    {".typ"},
	HTML:     {".html", ".htm"},
}

// tool is the converter each format needs on PATH.
// Candidate converter commands per format, tried in order.
//
// AsciiDoc lists asciidoctor.js as well as the Ruby implementation. The npm
// package installs as "asciidoctor" too, so this is for the case where both
// exist or where only the JavaScript one does under its qualified name; the
// flags used here are common to both.
//
// Markdown is deliberately absent: it is rendered in-process, so publishing
// Markdown needs nothing installed.
var tools = map[Format][]string{
	AsciiDoc: {"asciidoctor", "asciidoctor.js"},
	Typst:    {"typst"},
}

// Detect works out the source language from what is in the directory.
// HTML is only chosen when nothing else is present, so a directory holding
// both sources and a previous build is still recognised by its sources.
func Detect(dir string) (Format, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	found := map[Format]int{}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		for f, exts := range extensions {
			for _, x := range exts {
				if ext == x {
					found[f]++
				}
			}
		}
	}
	for _, f := range []Format{AsciiDoc, Markdown, Typst, HTML} {
		if found[f] > 0 {
			return f, nil
		}
	}
	return "", fmt.Errorf("%s: no .adoc, .md, .typ or .html files found", dir)
}

// navName is the navigation source for a format.
func navName(f Format) string {
	switch f {
	case AsciiDoc:
		return "nav.adoc"
	case Markdown:
		return "nav.md"
	case Typst:
		return "nav.typ"
	default:
		return "nav.html"
	}
}

// Options controls a pack run.
type Options struct {
	Source string
	OutDir string
	Format Format
	// Tool overrides the converter command for the detected format. It may
	// carry arguments -- "bundle exec asciidoctor", or a docker invocation --
	// which are placed before the ones the tool supplies.
	Tool    string
	Verbose func(string, ...any)
}

// command is a converter invocation: the program and any leading arguments.
type command []string

// splitCommand splits a command line on spaces, honouring single and double
// quotes so a path containing a space can be given.
func splitCommand(s string) command {
	var out command
	var cur strings.Builder
	var quote rune
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t':
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return out
}

// converter resolves the command for a format, honouring an override.
func converter(f Format, override string) (command, error) {
	if override != "" {
		cmd := splitCommand(override)
		if len(cmd) == 0 {
			return nil, fmt.Errorf("--tool is empty")
		}
		if _, err := exec.LookPath(cmd[0]); err != nil {
			return nil, fmt.Errorf("--tool %q: %w", cmd[0], err)
		}
		return cmd, nil
	}
	candidates := tools[f]
	if len(candidates) == 0 {
		return nil, nil // nothing to run, e.g. markdown or html
	}
	for _, name := range candidates {
		if _, err := exec.LookPath(name); err == nil {
			return command{name}, nil
		}
	}
	return nil, fmt.Errorf("%s is needed to pack %s sources but was not found on PATH; "+
		"pass --tool to point at it", strings.Join(candidates, " or "), f)
}

// Build converts the source directory and writes a zip into OutDir, named
// from the identity the bundle itself declares. Returns the archive path.
func Build(o Options) (string, error) {
	if o.Verbose == nil {
		o.Verbose = func(string, ...any) {}
	}
	format := o.Format
	if format == "" {
		var err error
		if format, err = Detect(o.Source); err != nil {
			return "", err
		}
		o.Verbose("detected %s source", format)
	}
	tool, err := converter(format, o.Tool)
	if err != nil {
		return "", err
	}
	if o.Tool != "" {
		o.Verbose("using %s", strings.Join(tool, " "))
	}

	work, err := os.MkdirTemp("", "tobar-segais-pack-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(work)

	if err := convert(format, tool, o.Source, work, o.Verbose); err != nil {
		return "", err
	}
	if err := stripStyles(work); err != nil {
		return "", err
	}
	if err := copyAssets(o.Source, work); err != nil {
		return "", err
	}
	return zipDir(work, o.OutDir)
}

// typstBundleEntry finds a Typst source that emits a whole bundle by itself.
//
// Typst's bundle export lets one compilation write many files, each declared
// with #document("page.html")[...]. That suits Typst's native structure --
// chapters joined into one source -- while still producing a page per file,
// so a book does not have to be dismantled to be published.
//
// It applies only when there is no nav.typ, which is the marker for the
// page-per-file layout.
func typstBundleEntry(dir string) (string, bool) {
	if fileExists(filepath.Join(dir, navName(Typst))) {
		return "", false
	}
	for _, name := range []string{"book.typ", "main.typ"} {
		if p := filepath.Join(dir, name); fileExists(p) {
			return p, true
		}
	}
	return "", false
}

// convert runs the format's converter over every page and its navigation.
func convert(f Format, tool command, src, out string, log func(string, ...any)) error {
	if f == Typst {
		if entry, ok := typstBundleEntry(src); ok {
			log("bundle export from %s", filepath.Base(entry))
			if err := run(tool, "compile", "--features", "html,bundle",
				"--format", "bundle", entry, out+string(os.PathSeparator)); err != nil {
				return err
			}
			if !fileExists(filepath.Join(out, "nav.html")) {
				return fmt.Errorf(`%s: bundle export produced no nav.html; emit one with #document("nav.html")[...]`,
					filepath.Base(entry))
			}
			return nil
		}
	}

	nav := filepath.Join(src, navName(f))
	if _, err := os.Stat(nav); err != nil {
		if f == Typst {
			return fmt.Errorf("%s has no %s and no book.typ or main.typ: a bundle needs either a page per file with a navigation document, or one source emitting documents",
				src, navName(f))
		}
		return fmt.Errorf("no %s in %s: a bundle needs a navigation document", navName(f), src)
	}

	pages, err := sourceFiles(src, f)
	if err != nil {
		return err
	}
	for _, p := range pages {
		if filepath.Base(p) == navName(f) {
			continue
		}
		if err := convertPage(f, tool, p, out); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(p), err)
		}
	}
	log("converted %d pages", len(pages)-1)
	return convertNav(f, tool, src, nav, out)
}

// sourceFiles lists the pages to convert.
//
// A leading underscore marks a partial: a file that exists to be pulled into
// another rather than to be a page of its own. Shared setup, a preamble, a
// chapter meant to be included -- naming it "_common.typ" keeps it out of the
// navigation and out of the bundle.
func sourceFiles(dir string, f Format) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		for _, x := range extensions[f] {
			if ext == x {
				out = append(out, filepath.Join(dir, e.Name()))
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

func htmlName(p string) string {
	return strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)) + ".html"
}

// convertPage converts one page. Converters are invoked as plainly as they
// allow: whatever they wrap the content in -- heads, stylesheets, footers --
// is removed afterwards, rather than suppressed by a flag per tool.
func convertPage(f Format, tool command, src, out string) error {
	dst := filepath.Join(out, htmlName(src))
	switch f {
	case AsciiDoc:
		// A plain invocation. In full output the document title is already an
		// <h1> in the header, and the footer -- the reason -s was needed --
		// is removed afterwards along with any stylesheet.
		return run(tool, "-o", dst, src)
	case Markdown:
		return renderMarkdownPage(src, dst)
	case Typst:
		return run(tool, "compile", "--format", "html", "--features", "html", src, dst)
	case HTML:
		return copyFile(src, dst)
	}
	return fmt.Errorf("unsupported format %q", f)
}

// convertNav produces the navigation document. Unlike the pages this is a
// whole document, because its head is the point: it carries the slug and
// version that identify the bundle.
func convertNav(f Format, tool command, srcDir, nav, out string) error {
	dst := filepath.Join(out, "nav.html")
	switch f {
	case AsciiDoc:
		return run(tool, "-a", "docinfosubs=attributes", "-o", dst, nav)
	case Markdown:
		return renderMarkdownNav(nav, dst)
	case Typst:
		return run(tool, "compile", "--format", "html", "--features", "html", nav, dst)
	case HTML:
		return copyFile(nav, dst)
	}
	return fmt.Errorf("unsupported format %q", f)
}

// copyAssets brings across anything that is not source: images, stylesheets,
// and any metadata file the author wrote by hand.
func copyAssets(src, out string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil || rel == "." {
			return err
		}
		if info.IsDir() {
			if rel == "build" || strings.HasPrefix(rel, ".") || strings.HasPrefix(rel, "_") {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(out, rel), 0o755)
		}
		if strings.HasPrefix(filepath.Base(p), "_") {
			return nil
		}
		switch strings.ToLower(filepath.Ext(p)) {
		case ".adoc", ".asciidoc", ".md", ".markdown", ".typ", ".sh":
			return nil // source and build scripts do not belong in the bundle
		case ".html", ".htm":
			if filepath.Dir(rel) == "." {
				return nil // pages are produced by the converter
			}
		}
		if filepath.Base(p) == "docinfo.html" || filepath.Base(p) == "meta.html" {
			return nil
		}
		return copyFile(p, filepath.Join(out, rel))
	})
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = io.Copy(w, in)
	return err
}

func run(tool command, args ...string) error {
	if len(tool) == 0 {
		return fmt.Errorf("no converter configured")
	}
	full := append(append([]string{}, tool[1:]...), args...)
	cmd := exec.Command(tool[0], full...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s: %s", tool[0], msg)
	}
	return nil
}

// zipDir writes the built tree to a temporary archive. Naming it needs the
// identity, which is read back out of the archive itself, so the same
// resolution rules apply as when the server loads it.
func zipDir(dir, outDir string) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(outDir, ".packing-*.zip")
	if err != nil {
		return "", err
	}
	zw := zip.NewWriter(tmp)
	err = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		w, err := zw.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
	if err == nil {
		err = zw.Close()
	}
	tmp.Close()
	if err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}
