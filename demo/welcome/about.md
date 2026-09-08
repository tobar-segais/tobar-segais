# About this site

This site is built by the tool it documents. Nothing here is hand-written HTML.

## How it is made

Every push to `main` runs a workflow that:

1. builds an archive from each directory of source — `docs/` for the user
   manual, and one for each of the format manuals — with `tobar-segais bundle`;
2. copies in the 1.15 and 1.16 archives, which are the original Eclipse
   InfoCenter bundles from the 1.x releases and are served exactly as they
   were published;
3. generates the whole site from that directory with `tobar-segais build`,
   which writes every page, the assets and a search index;
4. publishes the result.

So the site is a static snapshot with no server behind it, and the same
templates render it as would render it live. What you can do here that a
served site could also do — navigate, search, switch versions — is doing it
without a server; what you cannot do is add an archive and watch it appear,
which is the one thing publishing gives up. See
[Static sites](../../tobar-segais-manual/latest/static.html) in the manual.

## The name

Tobar Segais is the well the river of knowledge rises from, and the salmon of
knowledge swims in that river. The current in the banner above runs right to
left, because the salmon at the source is swimming into it.
