#! /usr/bin/env bash

# Usage:
#   ./package_binaries.sh <build_root>

set -e -u -o pipefail

here="$(dirname "${BASH_SOURCE[0]}")"

build_root="${1:-}"
if [ -z "$build_root" ] ; then
  echo 'Usage:  ./package_binaries.sh <build_root>'
  exit 1
fi

version="$(cat "$build_root/VERSION" | sed 's/+/./g')"

build_one_architecture() {
  echo "Making .deb package for $debian_arch."
  dest="$build_root/deb/$debian_arch"
  mkdir -p "$dest/build/DEBIAN"

  # Copy the entire filesystem layout, binaries and all, into the debian build
  # folder.
  cp -rp "$build_root"/binaries/"$debian_arch"/* "$dest/build"

  # Installed-Size is a required field in the control file. Without it Ubuntu
  # users will see warnings.
  size="$(du --summarize --block-size=1024 "$dest" | awk '{print $1}')"

  dependencies="Depends: "

  # Debian control file
  cat "$here/control.template" \
    | sed "s/@@VERSION@@/$version/" \
    | sed "s/@@ARCHITECTURE@@/$debian_arch/" \
    | sed "s/@@SIZE@@/$size/" \
    | sed "s/@@DEPENDENCIES@@/$dependencies/" \
    > "$dest/build/DEBIAN/control"

  fakeroot dpkg-deb --build "$dest/build" "$dest/forgefs-$version-$debian_arch.deb"
}

export debian_arch=amd64
build_one_architecture

#export debian_arch=i386
#build_one_architecture
