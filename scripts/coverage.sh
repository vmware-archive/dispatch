#!/bin/bash
# Generate test coverage statistics for Go packages.
#
# Works around the fact that `go test -coverprofile` does not work
# with multiple packages, see https://code.google.com/p/go/issues/detail?id=6909
#
# Usage: script/coverage [--html]
#
#     --html        Create HTML report and open it in browser
#

set -e -o pipefail

workdir=.cover

profile="$workdir/cover.out"
html="$workdir/cover.html"
mode=atomic
dir=$(dirname $0)

# list any files (or patterns) to explicitly exclude from coverage
# you should have a pretty good reason before putting items here
exclude_files=(gen hack)

join() { local IFS="$1"; shift; echo "$*"; }

excludes=$(join "|" ${exclude_files[@]} | sed -e 's/\./\\./g')

generate_cover_data() {
    rm -rf "$workdir"
    mkdir "$workdir"

    for pkg in "$@"; do
        f="$workdir/$(echo $pkg | tr / -).cover"
        go test -v -covermode="$mode" -coverprofile="$f" "$pkg"
    done

    echo "mode: $mode" >"$profile"
    grep -h -v "^mode:" "$workdir"/*.cover | $dir/exclude_ignore.py | egrep -v "$excludes" >>"$profile"
}

generate_cover_data $(go list ./... | grep -v /vendor/ | grep -v integration )
go tool cover -html="$profile" -o="$html"
