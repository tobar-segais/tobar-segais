#!/usr/bin/env bash
#
# A `gh` that answers from a local git repository, for `test.sh`.
#
# `compute-version.sh` reads the GitHub API, so a test of it needs an API. This
# file is a MODEL of the endpoints that the driver calls, built from the
# repository in `GITS_TEST_REPO`. The test copies it to a directory called `gh`
# on the PATH, so the driver runs with no change of its own.
#
# The model gives the SHAPES that the driver reads, and the driver applies its
# own `jq` expressions to them, so a change to one of those expressions is
# tested here. The shapes come from the API of GitHub, and a change there is
# what this file cannot see; keep the JSON below beside the documentation.
#
# `GITS_TEST_GH_FAIL=1` makes every call fail, which is how the test proves that
# a call that fails stops the driver and does not go out as an empty answer.

set -euo pipefail

if [ "${GITS_TEST_GH_FAIL:-}" = "1" ]; then
    echo "gh: the API said no (this is the stub, and it fails on purpose)" >&2
    exit 1
fi

repo="${GITS_TEST_REPO:?the stub needs GITS_TEST_REPO}"

[ "${1:-}" = "api" ] || {
    echo "the stub knows \`gh api\` only, and it got \`${1:-}\`" >&2
    exit 2
}
endpoint="$2"
shift 2

jq_expr=""
while [ $# -gt 0 ]; do
    case "$1" in
        --jq)
            jq_expr="$2"
            shift 2
            ;;
        *) shift ;;
    esac
done

# The commits of one range, in the shape that `compare` gives them. `jq -Rs .`
# writes the JSON of one message, because a message holds newlines and quotation
# marks of its own.
commits() {
    local first=1 message
    printf '['
    while IFS= read -r -d '' message; do
        [ "$first" -eq 1 ] || printf ','
        first=0
        printf '{"commit":{"message":%s}}' "$(printf '%s' "$message" | jq -Rs .)"
    done < <(git -C "$repo" log -z --format='%B' "$1")
    printf ']'
}

answer() {
    case "$endpoint" in
        */git/matching-refs/tags/v)
            local first=1 name object type
            printf '['
            while read -r name object type; do
                [ "$first" -eq 1 ] || printf ','
                first=0
                printf '{"ref":"%s","object":{"sha":"%s","type":"%s"}}' \
                    "$name" "$object" "$type"
            done < <(git -C "$repo" for-each-ref \
                --format='%(refname) %(objectname) %(objecttype)' 'refs/tags/*')
            printf ']'
            ;;

        # An ANNOTATED tag points at a tag object. This endpoint follows it
        # through to the commit, and the driver must call it.
        */git/tags/*)
            printf '{"object":{"sha":"%s"}}' \
                "$(git -C "$repo" rev-parse "${endpoint##*/}^{commit}")"
            ;;

        */compare/*)
            local spec base head type
            spec="${endpoint##*/compare/}"
            base="${spec%%...*}"
            head="${spec##*...}"

            # THIS ENDPOINT TAKES COMMITS, AND IT REFUSES A TAG OBJECT.
            #
            # The tags are annotated, and an annotated tag points at a tag
            # object and not at a commit. The real API refuses the sha of a tag
            # object here, so the driver MUST call `git/tags/` first and follow
            # the tag object through to its commit.
            #
            # `git log <sha>..` does NOT refuse it: git peels a tag object for
            # you, in silence. A stub built on `git log` alone would therefore
            # answer a tag object correctly, the tests would then pass with the
            # dereference REMOVED from the driver, and the one behaviour that
            # this check exists to protect would not be protected at all.
            type="$(git -C "$repo" cat-file -t "$base" 2>/dev/null || true)"
            if [ "$type" != "commit" ]; then
                echo "gh: HTTP 404: Not Found (the base \`$base\` is not a commit)" >&2
                exit 1
            fi

            printf '{"commits":'
            commits "$base..$head"
            printf '}'
            ;;

        # The `?` is escaped, because a `?` in a pattern of `case` matches one
        # character of anything.
        */commits\?sha=*)
            commits "${endpoint##*sha=}"
            ;;

        *)
            echo "the stub does not know the endpoint \`$endpoint\`" >&2
            exit 2
            ;;
    esac
}

if [ -n "$jq_expr" ]; then
    answer | jq -r "$jq_expr"
else
    answer
fi
