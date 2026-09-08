# The navigation document

Every bundle needs a `nav.md`. It carries the tree of pages, and the identity
of the bundle in its front matter.

## The tree

A nested list of links:

```markdown
- [Introduction](index.md)
- Guides
  - [Installing](install.md)
  - [Upgrading](upgrade.md)
```

Links point at the `.md` **sources**, and are rewritten to the generated
`.html` pages during conversion. So the same file works as navigation and as
something you can follow in an editor or on a git host.

The rewriting happens on the parsed document rather than by editing text, so a
reference-style link, or one split across lines, is handled correctly.

An entry with no link, like `Guides`, becomes a heading that groups its
children without being a page itself.

## Identity

YAML front matter at the top of `nav.md`:

```yaml
---
slug: handbook
version: 1.0.0
title: The Handbook
---
```

None of it is required. A bundle with no front matter at all takes its slug
and version from the filename, so `handbook-1.0.0.zip` is `handbook` at
`1.0.0`.

`title`, `aliases` and `hidden` are optional too, and mean what they do
everywhere else.

The slug must be lowercase letters, digits and hyphens: it goes straight into
a URL. The version is compared using semantic versioning, so `2.0.0-rc1` sorts
below `2.0.0` and is never chosen as the latest.

Front matter on the other pages is ignored, so you can keep whatever your
editor or git host expects there.
