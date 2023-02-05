#! /usr/bin/env bash

set -euox pipefail

here="$(dirname "${BASH_SOURCE[0]}")"
build_root="${1:-/tmp/keybase_build_$(date +%Y_%m_%d_%H%M%S)}"

$here/build_binary.sh $build_root
$here/deb/package_binary.sh $build_root
