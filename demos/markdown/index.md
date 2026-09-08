# Introduction

This manual documents how to publish Markdown documentation to Tobar Segais.
It is itself written in Markdown and published that way.

## What the server needs

A bundle is a zip containing HTML pages and a navigation document.

## What you need installed

Nothing.

Markdown is rendered inside the tool, so unlike AsciiDoc and Typst there is no
external converter to install:

```sh
tobar-segais bundle ./docs --out ./content
```

Detection is by file extension, so a directory of `.md` files is recognised as
Markdown without being told.

## Which Markdown

[GitHub Flavored Markdown](https://github.github.com/gfm/), by way of
[goldmark](https://github.com/yuin/goldmark) — the same renderer Hugo uses.
That means CommonMark, plus tables, strikethrough, task lists and autolinks.

Definition lists and footnotes are enabled as well, which GFM itself does not
specify.
