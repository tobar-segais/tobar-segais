#!/usr/bin/env bash
#
# The rules that give the version number. THIS FILE HOLDS THE RULES, and it is
# the only file that holds them.
#
# It reads no git repository and it makes no network call. It takes the facts as
# text and it writes the answer as text, so a test can give it any history in a
# millisecond. `compute-version.sh` collects the facts from the GitHub API and
# calls it.
#
# The build holds NO copy of these rules. The release stamps the number of the
# tag into the binary with `-ldflags`, and a build with no stamp reports what
# the Go module system recorded; neither calculates a number.
#
# # The rules for the number
#
#   fix, perf, revert   the third number goes up      (0.5.3 -> 0.5.4)
#   feat                the second number goes up     (0.5.3 -> 0.6.0)
#   a break             see below
#   everything else     no release
#
# A break is a `!` before the `:` in the first line, or a `BREAKING CHANGE:` or
# `BREAKING-CHANGE:` line in the body.
#
# WHILE THE FIRST NUMBER IS 0, A BREAK MOVES THE SECOND NUMBER, NOT THE FIRST.
# A 0.x version says that the interface can still change, so an automatic move
# to 1.0.0 would say something that the author did not intend. The move to 1.0.0
# is the act of a person: they make the tag `v1.0.0` by hand.
#
# NO FILE IN THE REPOSITORY HOLDS A NUMBER, so the tags alone give the floor.
#
# # Usage
#
#   version-rules.sh previous <head-tag> <tags-on-stdin>
#
#       Gives the release before HEAD. `<head-tag>` is the release tag that HEAD
#       carries, or an empty string. stdin holds every tag, one on each line, in
#       any order.
#
#       Writes:  previous=<vX.Y.Z or empty>
#                is-release=<true|false>
#
#   version-rules.sh next <previous-tag> <head-tag> <messages-on-stdin>
#
#       Gives the version. stdin holds the commit messages since the release
#       `<previous-tag>`, with a ZERO BYTE after each one, because a message
#       holds newlines of its own. stdin is not read when `<head-tag>` is set.
#
#       Writes:  version=<X.Y.Z>          the release to make, EMPTY when none
#                tag=<vX.Y.Z>             the tag to make, EMPTY when none
#                bump=<major|minor|patch> EMPTY when no commit asks for a release
#                is-release=<true|false>  true when HEAD carries a release tag
#                previous=<vX.Y.Z>        the release before, EMPTY when none

set -euo pipefail

# The form of a release tag. A tag with another form is not a release: `v1.0`
# and `v1.0.0-rc1` name something else, and a release number has three numbers.
release_tag='^v[0-9]+\.[0-9]+\.[0-9]+$'

# Every release tag, highest first.
#
# THE ORDER MUST BE A VERSION ORDER AND NEVER A TEXT ORDER. As text `v0.9.0`
# comes after `v0.10.0`, the caller then takes `v0.9.0` as the last release, and
# the next release goes BACKWARDS.
#
# `sort -V` does this, and the `sort` of macOS does not always hold `-V`. This
# code therefore makes a key of three numbers of fixed width and sorts the text
# of that key, which every `sort` can do.
sort_tags() {
    grep -E "$release_tag" |
        awk -F'[v.]' '{ printf "%010d %010d %010d %s\n", $2, $3, $4, $0 }' |
        sort -r |
        awk '{ print $4 }'
}

previous_of() {
    local head_tag="$1" tags
    tags="$(sort_tags || true)"

    if [ -z "$head_tag" ]; then
        # HEAD carries no release, so the last release is the highest tag.
        printf '%s\n' "$tags" | head -1
        return 0
    fi

    # HEAD is a release already. The release before it is the next tag DOWN the
    # list, and not simply the highest other tag: a person can build an earlier
    # release again, and the notes must then name the release before THAT one.
    printf '%s\n' "$tags" |
        awk -v t="$head_tag" 'found { print; exit } $0 == t { found = 1 }'
}

mode="${1:?give the mode: previous or next}"

case "$mode" in
    previous)
        head_tag="${2-}"
        previous="$(previous_of "$head_tag")"
        echo "previous=$previous"
        if [ -n "$head_tag" ]; then echo "is-release=true"; else echo "is-release=false"; fi
        exit 0
        ;;
    next) ;;
    *)
        echo "the mode must be \`previous\` or \`next\`, and it was \`$mode\`." >&2
        exit 2
        ;;
esac

previous="${2-}"
head_tag="${3-}"

if [ -n "$head_tag" ]; then
    # A tag of another form names something that is not a release, and a release
    # must never take its number from it. Stop, and say what a release tag is.
    if ! grep -qE "$release_tag" <<<"$head_tag"; then
        echo "\`$head_tag\` is not a release tag, so it names no release." >&2
        echo "A release tag has the form \`vX.Y.Z\`, such as \`v1.0.0\`." >&2
        echo "Remove the tag that is wrong, and make one with that form." >&2
        exit 2
    fi

    # HEAD carries a release tag, so the number is that tag and nothing is
    # calculated. `tag` and `bump` stay empty: this commit needs no second tag,
    # so a caller that acts on `tag` cannot tag one commit twice.
    echo "version=${head_tag#v}"
    echo "tag="
    echo "bump="
    echo "is-release=true"
    echo "previous=$previous"
    exit 0
fi

base="${previous#v}"
base="${base:-0.0.0}"
IFS='.' read -r major minor patch <<<"$base"
major="${major:-0}"
minor="${minor:-0}"
patch="${patch:-0}"

bump=""
while IFS= read -r -d '' message; do
    subject="${message%%$'\n'*}"

    # A break: `feat!:` or `feat(scope)!:` in the first line, or a
    # `BREAKING CHANGE:` line in the body.
    #
    # THIS TEST DOES NOT READ THE TYPE, so a capital letter in the type does not
    # stop it: `Feat!: a thing` is a break. The test below DOES read the type,
    # and it compares it against `feat` in small letters, so `Feat: a thing`
    # asks for no release at all. `check-title.sh` refuses both forms, and it
    # says which of the two happens.
    if [[ "$subject" =~ ^[a-zA-Z]+(\([^\)]*\))?!: ]] \
        || grep -qE '^BREAKING[ -]CHANGE:' <<<"$message"; then
        bump="major"
        continue
    fi

    if [[ ! "$subject" =~ ^([a-zA-Z]+)(\([^\)]*\))?!?:\  ]]; then
        continue
    fi
    type="${BASH_REMATCH[1]}"

    case "$type" in
        feat)
            if [ "$bump" != "major" ]; then bump="minor"; fi
            ;;
        fix | perf | revert)
            if [ -z "$bump" ]; then bump="patch"; fi
            ;;
        *) ;;
    esac
done

if [ -z "$bump" ]; then
    echo "version="
    echo "tag="
    echo "bump="
    echo "is-release=false"
    echo "previous=$previous"
    exit 0
fi

case "$bump" in
    major)
        # See the note at the top: a 0.x version moves the second number.
        if [ "$major" -eq 0 ]; then
            minor=$((minor + 1))
            patch=0
        else
            major=$((major + 1))
            minor=0
            patch=0
        fi
        ;;
    minor)
        minor=$((minor + 1))
        patch=0
        ;;
    patch)
        patch=$((patch + 1))
        ;;
esac

next="$major.$minor.$patch"

echo "version=$next"
echo "tag=v$next"
echo "bump=$bump"
echo "is-release=false"
echo "previous=$previous"
