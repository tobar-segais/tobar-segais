package search

import (
	"strings"
	"unicode"
)

// Marker delimits a highlighted run in a snippet. The template escapes the
// snippet and then replaces these, so a match can be emphasised without ever
// putting unescaped document text into the page.
const (
	MarkStart = "\x00<"
	MarkEnd   = "\x00>"
)

const snippetWords = 34

// snippet returns a short window of text around the best cluster of query
// terms, with matches marked. Falls back to the opening of the document when
// nothing matches in the body, which happens when a page matched on title.
func snippet(text string, terms []string) string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r)
	})
	if len(words) == 0 {
		return ""
	}
	want := map[string]bool{}
	for _, t := range terms {
		want[t] = true
	}

	norm := make([]string, len(words))
	for i, w := range words {
		norm[i] = strings.ToLower(strings.TrimFunc(w, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		}))
	}

	// Pick the window containing the most matches.
	best, bestAt := -1, 0
	for i := range words {
		if i+snippetWords > len(words) && i != 0 {
			break
		}
		n := 0
		for j := i; j < i+snippetWords && j < len(words); j++ {
			if want[norm[j]] {
				n++
			}
		}
		if n > best {
			best, bestAt = n, i
		}
	}
	if best <= 0 {
		bestAt = 0
	}

	end := bestAt + snippetWords
	if end > len(words) {
		end = len(words)
	}
	var b strings.Builder
	if bestAt > 0 {
		b.WriteString("… ")
	}
	for i := bestAt; i < end; i++ {
		if i > bestAt {
			b.WriteByte(' ')
		}
		if want[norm[i]] {
			b.WriteString(MarkStart)
			b.WriteString(words[i])
			b.WriteString(MarkEnd)
		} else {
			b.WriteString(words[i])
		}
	}
	if end < len(words) {
		b.WriteString(" …")
	}
	return b.String()
}
