package pack

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Markdown is rendered in-process with goldmark rather than by shelling out,
// so publishing Markdown needs nothing installed. It also lets two things be
// done properly that an external converter could not:
//
//   - Navigation links are rewritten on the syntax tree, so a link split
//     across lines or written reference-style is handled correctly. A textual
//     substitution of ".md)" would miss both.
//   - Identity comes from YAML front matter, so Markdown carries its own slug
//     and version instead of needing a separate metadata file.
func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,            // tables, strikethrough, autolinks, task lists
			extension.DefinitionList, // used by the appendix
			extension.Footnote,
			meta.Meta, // YAML front matter
		),
		goldmark.WithParserOptions(
			// Section ids feed the server's "On this page" column.
			parser.WithAutoHeadingID(),
		),
	)
}

// mdLinkRewriter turns links to Markdown sources into links to the pages that
// will be generated from them.
type mdLinkRewriter struct{}

func (mdLinkRewriter) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		link, ok := n.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}
		dest := string(link.Destination)
		if strings.Contains(dest, "://") {
			return ast.WalkContinue, nil // leave external links alone
		}
		path, frag, hasFrag := strings.Cut(dest, "#")
		for _, ext := range []string{".md", ".markdown"} {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				path = strings.TrimSuffix(path[:len(path)-len(ext)], "") + ".html"
				break
			}
		}
		if hasFrag {
			path += "#" + frag
		}
		link.Destination = []byte(path)
		return ast.WalkContinue, nil
	})
}

// renderMarkdownPage writes a content fragment.
func renderMarkdownPage(src, dst string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	md := newMarkdown()
	if err := md.Convert(raw, &buf, parser.WithContext(parser.NewContext())); err != nil {
		return err
	}
	return os.WriteFile(dst, buf.Bytes(), 0o644)
}

// renderMarkdownNav writes the navigation document: the tree, plus a head
// carrying whatever the front matter declared.
func renderMarkdownNav(src, dst string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithASTTransformers(util.Prioritized(mdLinkRewriter{}, 100)),
		),
	)
	ctx := parser.NewContext()
	var body bytes.Buffer
	if err := md.Convert(raw, &body, parser.WithContext(ctx)); err != nil {
		return err
	}

	front := meta.Get(ctx)
	get := func(k string) string {
		if v, ok := front[k]; ok {
			return strings.TrimSpace(fmt.Sprint(v))
		}
		return ""
	}

	var head bytes.Buffer
	if t := get("title"); t != "" {
		fmt.Fprintf(&head, "<title>%s</title>\n", html.EscapeString(t))
	}
	for _, k := range []string{"slug", "version", "title", "hidden", "copyright"} {
		if v := get(k); v != "" {
			fmt.Fprintf(&head, `<meta name="tobar-segais.%s" content="%s">`+"\n",
				k, html.EscapeString(v))
		}
	}
	if aliases := get("aliases"); aliases != "" {
		fmt.Fprintf(&head, `<meta name="tobar-segais.aliases" content="%s">`+"\n",
			html.EscapeString(aliases))
	}

	out := "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n" +
		head.String() + "</head>\n<body>\n" + body.String() + "</body>\n</html>\n"
	return os.WriteFile(dst, []byte(out), 0o644)
}
