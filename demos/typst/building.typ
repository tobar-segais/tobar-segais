= Building and publishing

== One command

```sh
tobar-segais bundle ./docs --out ./content
```

The tool detects Typst from the file extensions, works out which layout you
are using, converts accordingly, and writes a zip named from the slug and
version the bundle declares.

== What the tool runs

For a page per file, each page and the navigation document separately:

```sh
typst compile --format html --features html page.typ page.html
```

For a book, one compilation that writes every file:

```sh
typst compile --features html,bundle --format bundle book.typ out/
```

`--features html` is required because HTML export is not yet enabled by
default, and `bundle` likewise for bundle export. The tool passes them for
you.

== Publishing

Write the bundle where the server is watching:

```sh
tobar-segais bundle ./docs --out /srv/docs
```

The server notices within moments. There is nothing to restart.

== Assets

Images and other files next to the sources are copied into the bundle with
their layout intact, so relative references from a page keep working.

Files whose names begin with an underscore are treated as partials and are not
published.
