= Caveats

Typst's HTML export is young. These are the things that bite.

== HTML export is experimental

Typst prints a warning on every compile saying so, and its authors reserve the
right to change the behaviour. Bundle export carries the same warning. Treat
Typst bundles as a preview rather than a production commitment.

== Headings shift down a level, and that is fine

The exporter reserves `h1` for the document title, so `=` in source becomes
`h2` in output, `==` becomes `h3`, and so on.

You cannot correct this in Typst: `#set heading(offset: -1)` fails to compile,
because offsets must be zero or more.

You do not need to. The server normalises headings when it renders a page, so
the shallowest heading on the page becomes `h1` and the rest follow. Write `=`
for the page title and `==` for sections, as you would for a PDF, and the
result matches every other format.

== Lists use a hyphen

`-` starts a list item. `*` is strong emphasis, so a line beginning with `*`
is read as an unclosed delimiter and fails to compile. This is easy to get
wrong coming from Markdown or AsciiDoc.

== Pages cannot inherit setup

In a page-per-file layout each page is compiled on its own, so a helper or a
`#set` rule defined in another file is not in scope. Import what the page
needs from a partial. See #link("multifile.html")[Books and multiple files].

== Not every element has an HTML form

Layout-oriented features have no meaning in HTML and are either dropped or
reported as unsupported. Typst is a typesetting language first; its HTML
output is a second target, not the primary one.

If a construct does not survive, that is a limitation of the exporter rather
than of the bundle format.
