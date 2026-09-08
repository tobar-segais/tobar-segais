= Books and multiple files

Typst joins files with `#include`, producing one document. A documentation
bundle wants a page per file, each with its own address.

There are two ways to bridge that, and Typst supports both.

== Keeping one source

Typst's *bundle export* lets a single compilation write several files, each
declared with `#document`:

```typst
#document("nav.html")[
  #html.elem("meta", attrs: (name: "tobar-segais.slug", content: "handbook"))
  #html.elem("meta", attrs: (name: "tobar-segais.version", content: "1.0.0"))
  #html.elem("title")[The Handbook]

  - #link("index.html")[Introduction]
  - #link("chapter-two.html")[The second chapter]
]

#document("index.html")[
  = Introduction
  ...
]
```

Name that file `book.typ` or `main.typ`, leave out `nav.typ`, and the tool
uses bundle export:

```sh
tobar-segais bundle ./book --out ./content
  detected typst source
  bundle export from book.typ
```

This suits a book that already exists as one source. Chapters can stay in
separate files and be pulled in with `#import`, while each `#document` call
decides what becomes a page.

The navigation document is emitted the same way, so identity and tree still
come from the source.

== A page per file

The alternative is one `.typ` per page and a `nav.typ` listing them, as this
manual uses. The tool picks this when a `nav.typ` is present.

Each page is compiled on its own, so it cannot rely on setup performed by a
parent. This fails:

```typst
= Third chapter

#note[A helper defined elsewhere.]
```

with `error: unknown variable: note`. Import what the page needs:

```typst
#import "_common.typ": note
```

== Partials

A file whose name begins with an underscore is a partial: pulled into other
files, never a page of its own. Put shared setup in `_common.typ` and it stays
out of the navigation and out of the bundle.

The same convention applies to every format.

== Which to choose

Use bundle export for something that is a book: written as one work, read in
order, where the chapters share setup.

Use a page per file for something that is a manual: pages that are written,
reviewed and linked to separately.

== What replaces the outline

`#outline()` builds a table of contents from headings and links to anchors
within one document. In a bundle the server provides that instead: the
navigation tree spans pages and versions, and the *On this page* column is
built from each page's own headings.
