= The navigation document

Every bundle needs a navigation document. It carries the tree of pages and the
identity of the bundle.

In a page-per-file layout it is `nav.typ`. With bundle export it is whatever
your source emits as `nav.html`. The content is the same either way.

== The tree

Write it as a list of links. Nesting becomes nesting:

```typst
- #link("index.html")[Introduction]
- Guides
  - #link("install.html")[Installing]
  - #link("upgrade.html")[Upgrading]
```

Links point at the generated `.html` files, not at the `.typ` sources. Typst
has no cross-reference form that rewrites extensions, so this is one place
where Typst is more manual than AsciiDoc.

An entry with no link, like `Guides`, becomes a heading that groups its
children without being a page itself.

== Identity

Slug and version are declared as HTML meta elements, and the title as a title
element:

```typst
#html.elem("meta", attrs: (name: "tobar-segais.slug", content: "handbook"))
#html.elem("meta", attrs: (name: "tobar-segais.version", content: "1.0.0"))
#html.elem("title")[The Handbook]
```

Typst emits these into the document body rather than its head. The server
reads meta elements wherever they appear, so this works, but it is worth
knowing if you inspect the generated file and wonder why the head looks bare.

Neither is required. A bundle with no metadata at all takes its slug and
version from the filename, so `handbook-1.0.0.zip` is `handbook` at `1.0.0`.

The slug must be lowercase letters, digits and hyphens: it goes straight into
a URL. The version is compared using semantic versioning.
