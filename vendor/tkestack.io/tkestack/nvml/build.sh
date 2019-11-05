#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail


ROOT=$(dirname "$BASH_SOURCE[0]")
ROOT=$(cd $ROOT; pwd -P)

MKDIR=$(which mkdir)
LN=$(which ln)

GOPATH="${ROOT}/go/"

function nvml::prepare() {
  ${MKDIR} -p "${GOPATH}/src/tkestack.io/tkestack/"
  $LN -sf $(cd "${ROOT}/" && pwd -P) "${GOPATH}/src/tkestack.io/tkestack/"
}

function nvml::build() {
  (
    export GOPATH=${GOPATH}
    go build -o "${ROOT}/go-nvml" "./examples/"
  )
}


nvml::prepare
nvml::build
