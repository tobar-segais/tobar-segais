package search

import (
	"context"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/net/html"

	"github.com/tobar-segais/tobar-segais/internal/bundle"
)

// Indexer keeps a published Index in step with a catalogue, rebuilding in the
// background so that serving never waits on it.
type Indexer struct {
	log   *slog.Logger
	idx   atomic.Pointer[Index]
	prog  Progress
	queue chan []*bundle.Bundle
}

// NewIndexer returns an indexer with an empty published index, ready to serve
// "not indexed yet" answers immediately.
func NewIndexer(log *slog.Logger) *Indexer {
	ix := &Indexer{log: log, queue: make(chan []*bundle.Bundle, 8)}
	ix.idx.Store(NewBuilder().Freeze())
	return ix
}

// Index returns the currently published index.
func (ix *Indexer) Index() *Index { return ix.idx.Load() }

// Status reports build progress.
func (ix *Indexer) Status() Status { return ix.prog.Status() }

// Rebuild asks for the catalogue to be indexed. It never blocks the caller:
// a full queue means a rebuild is already pending, which supersedes this one.
func (ix *Indexer) Rebuild(all []*bundle.Bundle) {
	select {
	case ix.queue <- all:
	default:
	}
}

// Run processes rebuild requests until ctx is done.
func (ix *Indexer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case all := <-ix.queue:
			ix.build(ctx, all)
		}
	}
}

func (ix *Indexer) build(ctx context.Context, all []*bundle.Bundle) {
	start := time.Now()
	b := NewBuilder()

	var total int64
	pages := make([]page, 0, 256)
	for _, bn := range all {
		// Titles come from the table of contents where it names one; the
		// root page takes the table of contents label.
		labels := map[string]string{}
		if h := clean(bn.TOC.Href); h != "" {
			labels[h] = bn.TOC.Label
		}
		bn.TOC.Walk(func(t *bundle.Topic) {
			if h := clean(t.Href); h != "" {
				if _, ok := labels[h]; !ok {
					labels[h] = t.Label
				}
			}
		})
		for _, href := range bn.TOC.Pages() {
			pages = append(pages, page{b: bn, href: href, title: labels[href]})
			total++
		}
	}
	ix.prog = Progress{}
	ix.prog.AddTotal(total)
	ix.prog.SetReady(false)

	for _, p := range pages {
		if ctx.Err() != nil {
			return
		}
		raw, err := p.b.Archive.ReadFile(p.href)
		if err != nil {
			ix.log.Debug("indexer: unreadable page", "bundle", p.b.Slug, "href", p.href, "err", err)
			ix.prog.Step()
			continue
		}
		text, title := extract(raw)
		if p.title != "" {
			title = p.title
		}
		b.Add(Doc{
			Slug:    p.b.Slug,
			Version: p.b.Version,
			Href:    p.href,
			Title:   title,
		}, text)
		ix.prog.Step()
	}

	idx := b.Freeze()
	ix.idx.Store(idx)
	ix.prog.SetReady(true)
	ix.log.Info("index rebuilt",
		"documents", idx.Len(), "bundles", len(all), "took", time.Since(start).Round(time.Millisecond))
}

// clean strips a fragment or query from an href.
func clean(href string) string {
	href = strings.TrimSpace(href)
	if i := strings.IndexAny(href, "#?"); i != -1 {
		href = href[:i]
	}
	return href
}

type page struct {
	b     *bundle.Bundle
	href  string
	title string
}

// extract pulls readable text and the document title out of an HTML page,
// skipping script and style content.
func extract(raw []byte) (text, title string) {
	doc, err := html.Parse(strings.NewReader(string(raw)))
	if err != nil {
		return "", ""
	}
	var sb strings.Builder
	var inTitle bool
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "noscript":
				return
			case "title":
				inTitle = true
				defer func() { inTitle = false }()
			}
		}
		if n.Type == html.TextNode {
			if inTitle {
				title += n.Data
			} else if s := strings.TrimSpace(n.Data); s != "" {
				sb.WriteString(s)
				sb.WriteByte(' ')
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return sb.String(), strings.TrimSpace(title)
}
