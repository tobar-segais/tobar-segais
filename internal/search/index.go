// Package search builds a full-text index over bundle content.
//
// The index is built in the background and published by pointer swap, so the
// server answers navigation and content requests from the moment it starts
// and search becomes available when it is ready. Queries against a partial
// index are answered honestly rather than returning a misleadingly empty
// result set.
package search

import (
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
)

// Doc is one indexed page.
type Doc struct {
	Slug    string
	Version string
	Href    string
	Title   string
	text    string // kept for snippets
}

type posting struct {
	doc  int32
	freq int32
}

// Index is an immutable inverted index. Build one with a Builder.
type Index struct {
	docs  []Doc
	terms map[string][]posting
}

// Result is a ranked hit with a highlighted snippet.
type Result struct {
	Doc
	Score   float64
	Snippet string
}

// Tokenise splits text the same way for indexing and querying: runs of
// letters and digits, lowercased. Keeping one function means a query can
// never tokenise differently from the text it is searching.
func Tokenise(s string) []string {
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			out = append(out, b.String())
			b.Reset()
		}
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		} else {
			flush()
		}
	}
	flush()
	return out
}

// Builder accumulates documents and freezes them into an Index.
type Builder struct {
	mu    sync.Mutex
	docs  []Doc
	terms map[string][]posting
}

// NewBuilder starts an empty index.
func NewBuilder() *Builder {
	return &Builder{terms: map[string][]posting{}}
}

// Add indexes one page. Title terms are weighted by counting them twice,
// which is crude but keeps the ranking function simple and predictable.
func (b *Builder) Add(d Doc, body string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	d.text = body
	id := int32(len(b.docs))
	b.docs = append(b.docs, d)

	freq := map[string]int32{}
	for _, t := range Tokenise(d.Title) {
		freq[t] += 2
	}
	for _, t := range Tokenise(body) {
		freq[t]++
	}
	for t, n := range freq {
		b.terms[t] = append(b.terms[t], posting{doc: id, freq: n})
	}
}

// Freeze returns an immutable index over everything added so far.
func (b *Builder) Freeze() *Index {
	b.mu.Lock()
	defer b.mu.Unlock()
	docs := make([]Doc, len(b.docs))
	copy(docs, b.docs)
	terms := make(map[string][]posting, len(b.terms))
	for t, p := range b.terms {
		cp := make([]posting, len(p))
		copy(cp, p)
		terms[t] = cp
	}
	return &Index{docs: docs, terms: terms}
}

// Len is the number of indexed documents.
func (i *Index) Len() int { return len(i.docs) }

// Search returns hits ordered by score. All query terms must be present:
// documentation search is a lookup, not a discovery engine, and partial
// matches mostly produce noise.
func (i *Index) Search(query string, limit int) []Result {
	terms := Tokenise(query)
	if len(terms) == 0 || len(i.docs) == 0 {
		return nil
	}

	scores := map[int32]float64{}
	counts := map[int32]int{}
	n := float64(len(i.docs))
	for _, t := range terms {
		postings := i.terms[t]
		if len(postings) == 0 {
			return nil // an absent term means no document satisfies the query
		}
		idf := math.Log(1 + n/float64(len(postings)))
		for _, p := range postings {
			scores[p.doc] += (1 + math.Log(float64(p.freq))) * idf
			counts[p.doc]++
		}
	}

	out := make([]Result, 0, len(scores))
	for id, s := range scores {
		if counts[id] != len(terms) {
			continue
		}
		d := i.docs[id]
		out = append(out, Result{Doc: d, Score: s, Snippet: snippet(d.text, terms)})
	}
	sort.Slice(out, func(a, b int) bool {
		if out[a].Score != out[b].Score {
			return out[a].Score > out[b].Score
		}
		return out[a].Title < out[b].Title
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// ExportDoc is an indexed page, flattened for a client-side search index.
type ExportDoc struct {
	Slug    string `json:"s"`
	Version string `json:"v"`
	Href    string `json:"h"`
	Title   string `json:"t"`
	Text    string `json:"x"`
}

// Export returns every indexed document. A generated site ships these and
// searches them in the browser, since there is no server to ask.
func (i *Index) Export() []ExportDoc {
	out := make([]ExportDoc, 0, len(i.docs))
	for _, d := range i.docs {
		out = append(out, ExportDoc{
			Slug: d.Slug, Version: d.Version, Href: d.Href, Title: d.Title, Text: d.text,
		})
	}
	return out
}

// Status reports indexing progress so the UI can say what is happening
// rather than pretending an empty result set is an answer.
type Status struct {
	Done  int64
	Total int64
	Ready bool
}

// Progress is a live, concurrency-safe view of an in-flight build.
type Progress struct {
	done  atomic.Int64
	total atomic.Int64
	ready atomic.Bool
}

func (p *Progress) AddTotal(n int64) { p.total.Add(n) }
func (p *Progress) Step()            { p.done.Add(1) }
func (p *Progress) SetReady(b bool)  { p.ready.Store(b) }

// Status snapshots progress.
func (p *Progress) Status() Status {
	return Status{Done: p.done.Load(), Total: p.total.Load(), Ready: p.ready.Load()}
}
