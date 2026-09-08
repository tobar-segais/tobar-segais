# Tobar Segais

[![Site][LogoImage]][website]

Serves versioned documentation archives. Point it at a directory of `.zip` or
`.jar` files and it publishes each one at a stable URL, with a navigation tree,
full-text search and a version picker.

It is a single static binary with no runtime dependencies: no application
server, no JVM, no database.

The manual, and a live demonstration, are at **[tobarsegais.org][website]**.

## Install

```console
$ go install github.com/tobar-segais/tobar-segais/cmd/tobar-segais@latest
```

Binaries for Linux, macOS and Windows are attached to [each release][releases],
and the image is on the GitHub registry:

```console
$ docker run --rm -p 8080:8080 -v "$PWD/content:/content" \
    ghcr.io/tobar-segais/tobar-segais:latest
```

## Use

Serve a directory of archives:

```console
$ tobar-segais serve --content ./content
```

Build an archive from a directory of Markdown, AsciiDoc or Typst source:

```console
$ tobar-segais bundle ./docs --out ./content
```

Generate the whole site as static files, for GitHub Pages or any other host:

```console
$ tobar-segais build --content ./content --out ./site
```

Markdown needs nothing installed. AsciiDoc needs `asciidoctor` and Typst needs
`typst`; the manual says how to point at either.

## 2.0

2.0 is a rewrite in Go. The 1.x releases were a Java web application that ran
in a servlet container and read its content from inside its own WAR file, and
that source is in the history of this repository.

The 1.x bundles still work: an Eclipse InfoCenter help bundle needs no Eclipse
and no OSGi to be served, which was the point of the project in the first
place. The demo serves the original 1.15 and 1.16 manuals unchanged, beside the
2.0 one.

What changed:

- **Content is no longer packaged with the application.** Archives live in a
  directory the server watches, and adding or removing one takes effect without
  a restart.
- **A bundle no longer has to be an Eclipse bundle.** A zip with some HTML and
  a generated `nav.html` is enough, and `tobar-segais bundle` makes one from
  Markdown, AsciiDoc or Typst.
- **The site can be static.** `tobar-segais build` writes every page and a
  search index that is queried in the browser, so there is nothing to run.

## Build from source

```console
$ go build ./cmd/tobar-segais
$ go test ./...
```

## Releases

Every merge to `main` is considered for a release. The pull request title is
the commit message, it follows [Conventional Commits][cc], and it decides the
number: `fix:` moves the patch, `feat:` moves the minor, and a `!` or a
`BREAKING CHANGE:` moves the major. A commit that asks for nothing — `docs:`,
`chore:` — releases nothing. No file in the repository holds a version number;
the tag does.

## Licence

Apache License 2.0. See [LICENSE](LICENSE).

[LogoImage]: https://tobarsegais.org/static/tobarsegais.png
[website]: https://tobarsegais.org/
[releases]: https://github.com/tobar-segais/tobar-segais/releases
[cc]: https://www.conventionalcommits.org/
