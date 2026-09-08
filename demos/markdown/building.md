# Building and publishing

## One command

```sh
tobar-segais bundle ./docs --out ./content
```

That renders every `.md` page, renders `nav.md` into the navigation document,
and writes a zip named from the slug and version the front matter declares.

## What happens

Pages are rendered as fragments, because the server supplies the surrounding
page: its own layout, the navigation tree, the version pills and search.

The navigation document is rendered as a whole document, because its head is
where the identity goes.

Section ids are generated automatically from headings, and the server uses
them to build the *On this page* column.

## Publishing

Write the bundle where the server is watching:

```sh
tobar-segais bundle ./docs --out /srv/docs
```

The server notices within moments. There is nothing to restart.

## Assets

Images and other files next to the sources are copied into the bundle with
their layout intact, so relative references from a page keep working.
