#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail


ROOT=$(dirname "$BASH_SOURCE[0]")
ROOT=$(cd $ROOT; pwd -P)

function nvml::build() {
  go build -o "${ROOT}/go-nvml" "./examples/"
}

nvml::build
