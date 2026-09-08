# Appendix: Markdown elements

An illustration of how GitHub Flavored Markdown reaches the page.

## Inline

Ordinary prose, with *emphasis*, **strong text**, ~~strikethrough~~,
`inline code` and [links](https://github.github.com/gfm/).

Autolinks work too: https://github.com/yuin/goldmark

## Lists

- A bullet
- Another
  - Nested one level
  - And another

1. Numbered
2. Second

Task lists, from GFM:

- [x] Render Markdown without an external tool
- [x] Rewrite navigation links on the syntax tree
- [ ] Anything else

## Code

```sh
tobar-segais serve --content ./content --addr :8080
```

## Quotes

> A documentation server should get out of the way of the documentation.

## Tables

| Flag        | Default     | Meaning               |
|-------------|-------------|-----------------------|
| `--content` | `./content` | Directory of archives |
| `--addr`    | `:8080`     | Listen address        |
| `--poll`    | `30s`       | Rescan interval       |

## Definition lists

Bundle
: A zip of HTML pages and a navigation document.

Slug
: The name a bundle is served under.

## Footnotes

Markdown is rendered in-process[^1].

[^1]: With goldmark, the renderer Hugo uses.
