# Caveats

Markdown is the least troublesome of the three formats, because nothing is
shelled out to.

## Raw HTML is passed through, then filtered

Markdown permits inline HTML, and it survives into the page. The server strips
`<script>` elements and inline event handlers when it renders content, so a
help page cannot run code in the container's origin.

Do not rely on scripts in documentation; they will not run.

## Front matter belongs on the navigation document

`slug` and `version` are read from `nav.md` only. Front matter on other pages
is parsed and ignored, which is deliberate: a page cannot rename the bundle
it happens to sit in.

## Only relative links are rewritten

Links to `.md` files are rewritten to `.html`. Anything with a scheme is left
exactly as written, so external links are never touched.

## Nesting is CommonMark's

A sub-list is indented to line up with its parent's content — two spaces after
a `-` marker. This is CommonMark's rule, and it is not what every Markdown
tool does; some require four.

If a group's children are missing from the sidebar, check the indentation.
