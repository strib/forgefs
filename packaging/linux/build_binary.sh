#! /usr/bin/env bash

set -euox pipefail

here="$(dirname "${BASH_SOURCE[0]}")"
this_repo="$(git -C "$here" rev-parse --show-toplevel ||
  echo -n "$GOPATH/src/github.com/strib/forgefs")"

# Take the first argument as the build root, or a tmp dir if there is no
# second argument. Absolutify the build root, because we cd around in this
# script, and also because GOPATH is not allowed to be relative.
build_root="${1:-/tmp/keybase_build_$(date +%Y_%m_%d_%H%M%S)}"
mkdir -p "$build_root"
build_root="$(realpath "$build_root")"

# Record the version now, and write it to the build root. Because it
# uses a timestamp, it's important that other scripts use this file
# instead of recomputing the version themselves.
version="$("$here/../version.sh" "$@")"
echo -n "$version" > "$build_root/VERSION"

echo "Building version $version in $build_root"

build_one_architecture() {
  layout_dir="$build_root/binaries/$debian_arch"
  mkdir -p "$layout_dir/usr/bin"

  # Assemble a custom GOPATH.
  export GOPATH="$build_root/gopaths/$debian_arch"
  mkdir -p "$GOPATH/src/github.com/strib"
  ln -snf "$this_repo" "$GOPATH/src/github.com/strib/forgefs"

  # Build the client binary. Note that `go build` reads $GOARCH.
  echo "Building forgefs for $GOARCH..."
  go build -o \
     "$layout_dir/usr/bin/forgefs" github.com/strib/forgefs/forgefs
}

echo "forgefs: Building for x86-64"
export GOARCH=amd64
export debian_arch=amd64
build_one_architecture

echo "forgefs: Building for x86"
export GOARCH=386
export debian_arch=i386
build_one_architecture

