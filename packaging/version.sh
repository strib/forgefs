#! /usr/bin/env bash

set -e -u -o pipefail

here="$(dirname "$BASH_SOURCE")"

version_file="$(dirname "$BASH_SOURCE")/../version.go"
version="$(cat "$version_file" | grep 'Version =' | grep -oE '[0-9]+(.[0-9]+)+')"
build=""

current_date="$(date -u +%Y%m%d%H%M%S)" # UTC
commit_short="$(git -C "$here" log -1 --pretty=format:%h || \
    echo -n ${SOURCE_COMMIT:0:10})"
build="-$current_date+$commit_short"

echo "$version$build"
