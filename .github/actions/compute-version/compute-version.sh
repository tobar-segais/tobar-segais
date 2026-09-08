#!/usr/bin/env bash
#
# Gives the version of a commit from the GITHUB API, with no clone of the
# history.
#
# The action beside this file runs it. The rules are in `version-rules.sh`; this
# file collects the facts only.
#
# # Why the API and not a clone
#
# This runs on every push to `main` and on every pull request, and its only work
# is to answer "is there a release, and what is the number?". A clone with
# `fetch-depth: 0` fetches the whole history to answer that. Four API calls
# answer it in about a second, whatever the size of the history.
#
# A workflow still needs `actions/checkout` to get the files of this action, but
# the DEFAULT checkout is enough. Never ask for `fetch-depth: 0` for this.
#
# # Which tag is the last release
#
# The tags of this repository sit on the commits of `main`. The last release is
# the highest tag BY VERSION, and never the nearest tag in the history and
# never the highest as text: as text `v0.9.0` comes after `v0.10.0`, and the
# next release then goes backwards. `version-rules.sh` does the sorting. The
# commit of that tag is the start of the range of messages to read, and the
# range excludes the commit itself, because its message was in the range of the
# release before.
#
# # It must fail LOUDLY
#
# A call to the API can fail for a reason that has nothing to do with this
# project: a rate limit, a 5xx, a token that expired. This file then writes
# nothing and exits with a value that is not zero. The caller MUST NOT swallow
# that: a release that goes green and releases nothing is the fault that this
# project refuses to accept. The action beside this file writes the output to a
# file and does not pipe it, so the exit value of this script is the exit value
# of the step.
#
# # Usage
#
#   compute-version.sh <repository> <head-sha> [<head-tag>]
#
# `<head-tag>` is the release tag that the commit carries, and it comes from the
# event: a push of a tag names it in `github.ref_name`. Give an empty string for
# a push to a branch. This costs nothing, and it saves one API call for each tag
# that the repository holds.
#
# It needs `gh` and `jq`, which every GitHub runner holds, and a token in
# `GH_TOKEN`.

set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
rules="$here/version-rules.sh"

repository="${1:?give the repository, such as stephenc/tobar-segais}"
head_sha="${2:?give the sha of the commit}"
head_tag="${3-}"

# Every `v*` tag in one call. The order that the API gives is NOT a version
# order, so `version-rules.sh` sorts them. Each line is the name of a ref, the
# sha of the object that it points at, and the type of that object.
refs="$(gh api "repos/$repository/git/matching-refs/tags/v" --paginate \
    --jq '.[] | [(.ref | sub("^refs/tags/"; "")), .object.sha, .object.type] | @tsv')"

previous="$(printf '%s\n' "$refs" | cut -f1 | "$rules" previous "$head_tag" |
    sed -n 's/^previous=//p')"

if [ -n "$head_tag" ]; then
    # A release needs no commit message: the tag gives the number.
    "$rules" next "$previous" "$head_tag" </dev/null
    exit 0
fi

messages() {
    if [ -z "$previous" ]; then
        # No release exists, so every commit counts. This happens once, before
        # the first release.
        gh api "repos/$repository/commits?sha=$head_sha" --paginate |
            jq -j '.[] | .commit.message + "\u0000"'
        return 0
    fi

    local sha type commit
    sha="$(printf '%s\n' "$refs" | awk -F'\t' -v t="$previous" '$1 == t { print $2; exit }')"
    type="$(printf '%s\n' "$refs" | awk -F'\t' -v t="$previous" '$1 == t { print $3; exit }')"

    # AN ANNOTATED TAG POINTS AT A TAG OBJECT, AND NOT AT A COMMIT. Our tags are
    # annotated, so this step is the one that works in a test with lightweight
    # tags and fails on the real repository. Follow the tag object through to
    # the commit before comparing against it.
    if [ "$type" = "tag" ]; then
        commit="$(gh api "repos/$repository/git/tags/$sha" --jq '.object.sha')"
    else
        commit="$sha"
    fi

    # `compare` gives the commits and their messages in one call, and the range
    # EXCLUDES the base, so the message of the tagged commit itself is not read
    # again. It returns 250 commits at the most; a release with more commits
    # than that gives a number from the last 250, which is more than this
    # project makes between two releases.
    #
    # `jq -j` writes no newline of its own, so the zero byte is the only
    # separator. A commit message holds newlines, and a newline separator would
    # make one message look like several.
    gh api "repos/$repository/compare/$commit...$head_sha" |
        jq -j '.commits[] | .commit.message + "\u0000"'
}

messages | "$rules" next "$previous" ""
