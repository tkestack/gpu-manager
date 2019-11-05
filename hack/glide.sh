#!/usr/bin/env bash
set -o errexit
set -o pipefail
set -o nounset

ROOT=$(cd $(dirname ${BASH_SOURCE[0]})/.. && pwd -P)

source "${ROOT}/hack/common.sh"

GLIDE="${GLIDE:-glide}"

(
  plugin::setup_env
  ${GLIDE} up --strip-vendor --no-recursive --skip-test
)
