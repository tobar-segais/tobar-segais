#!/usr/bin/env bash
#
# Every local link in a generated site lands on a file that exists.
#
# The generator writes links from the catalogue, from the navigation tree and
# from the version picker, and each of those is built from a different source:
# the picker in particular offers a page in another version, which may not hold
# that page. Twice already a link like that has reached a reader as a bare 404,
# and both times the site itself was the only place it showed.
#
#   check-links.sh <site-directory>
#
# It reads the files and makes no network call, so an external link is out of
# scope here: this asks whether the site is consistent with itself.

set -euo pipefail

site="${1:?usage: check-links.sh <site-directory>}"
cd "$site"

fail=0
checked=0

# href and src, with the value in group 1. A fragment or a query is cut off
# below: the file is what has to exist.
while IFS= read -r page; do
    dir="$(dirname "$page")"
    while IFS= read -r target; do
        case "$target" in
            ''|'#'*|http://*|https://*|mailto:*|data:*|//*) continue ;;
        esac
        target="${target%%#*}"
        target="${target%%\?*}"
        [ -n "$target" ] || continue

        case "$target" in
            /*) path=".$target" ;;
            *)  path="$dir/$target" ;;
        esac
        # A link to a directory is a link to its index page.
        case "$path" in
            */) path="${path}index.html" ;;
        esac

        checked=$((checked + 1))
        if [ ! -e "$path" ]; then
            echo "$page -> $target" >&2
            fail=$((fail + 1))
        fi
    done < <(grep -oE '(href|src)="[^"]*"' "$page" | sed -E 's/^(href|src)="//; s/"$//')
done < <(find . -name '*.html' -type f)

if [ "$fail" -gt 0 ]; then
    echo >&2
    echo "$fail link(s) in the generated site have nowhere to land." >&2
    exit 1
fi

echo "$checked local links, all of which land somewhere."
