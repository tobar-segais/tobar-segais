= Introduction

This manual documents how to publish #link("https://typst.app")[Typst]
documentation to Tobar Segais. It is itself written in Typst and published
that way, so what you are reading is the output of the process it describes.

== What the server needs

A bundle is a zip containing HTML pages and a navigation document. Typst's
HTML export produces both, so no other tooling is involved.

== Two ways to lay out a manual

/ A page per file: One `.typ` per page, and a `nav.typ` listing them. This
  manual is built that way.

/ One source, many documents: A single `book.typ` or `main.typ` that emits
  each page with `#document`, using Typst's bundle export. This suits
  something written as a book.

Both are described in #link("multifile.html")[Books and multiple files]. The tool
chooses between them by looking for a `nav.typ`.

== What you need installed

Typst, on `PATH`. The server invokes it for you:

```sh
tobar-segais bundle ./docs --out ./content
```

Detection is by file extension, so a directory of `.typ` files is recognised
as Typst without being told. If your Typst is somewhere unusual, point at it
with `--tool`.
