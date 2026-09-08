# Tobar Segais

> An ordinary salmon ate nine hazelnuts that fell into Tobar Segais - the Well
> of Wisdom - from the nine hazel trees that surrounded the well, and gained
> all the world's knowledge.
>
> — Irish mythology

Tobar Segais serves versioned documentation archives. Point it at a directory
of `.zip` or `.jar` files and it publishes each one at a stable URL, with a
navigation tree, full-text search and a version picker.

It is a single static binary. There is no application server, no JVM and no
database.

## This site is the demo

Everything here is served by Tobar Segais, and every manual in the sidebar is
an archive in a directory:

- **[The user manual](../../tobar-segais-manual/latest/)** — how to publish
  documentation to it and how to run it. The version picker on it reaches the
  1.15 and 1.16 manuals, which are the original Eclipse InfoCenter bundles from
  the 1.x releases, served unchanged.
- **[Bundles from Markdown](../../markdown-bundles/latest/)**,
  **[from AsciiDoc](../../asciidoc-bundles/latest/)** and
  **[from Typst](../../typst-bundles/latest/)** — one manual for each source
  format, each documenting how to publish that format, and each published that
  way.

## Get it

```console
$ go install github.com/tobar-segais/tobar-segais/cmd/tobar-segais@latest
```

Or run the image:

```console
$ docker run --rm -p 8080:8080 -v "$PWD/content:/content" \
    ghcr.io/tobar-segais/tobar-segais:latest
```

Binaries for Linux, macOS and Windows are attached to
[each release](https://github.com/tobar-segais/tobar-segais/releases).

## Elsewhere

- [The source](https://github.com/tobar-segais/tobar-segais)
- [Follow @tobarsegais](https://twitter.com/tobarsegais)
