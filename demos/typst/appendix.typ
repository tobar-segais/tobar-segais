= Appendix: Typst elements

An illustration of how common Typst constructs reach the page. Everything
below is written in the source of this page.

== Inline

Ordinary prose, with #emph[emphasis], #strong[strong text],
`inline code` and #link("https://typst.app")[links].

== Lists

- A bullet
- Another
  - Nested one level
  - And another

+ Numbered
+ Second

== Code

```sh
tobar-segais serve --content ./content --addr :8080
```

== Quotes

#quote(block: true)[
  A documentation server should get out of the way of the documentation.
]

== Tables

#table(
  columns: 3,
  table.header([Flag], [Default], [Meaning]),
  [`--content`], [`./content`], [Directory of archives],
  [`--addr`], [`:8080`], [Listen address],
  [`--poll`], [`30s`], [Rescan interval],
)

== Terms

/ Bundle: A zip of HTML pages and a navigation document.
/ Slug: The name a bundle is served under.
