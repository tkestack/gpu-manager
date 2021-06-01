#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

ROOT=$(cd $(dirname ${BASH_SOURCE[0]})/.. && pwd -P)

source "${ROOT}/hack/common.sh"

function plugin::build() {
  (
    for arg; do
        case $arg in
        test)
            plugin::run_test
            ;;
        proto)
            plugin::generate_proto
            ;;
        img)
            plugin::generate_img
            ;;
        fmt)
            plugin::fmt
            ;;
        *)
            plugin::build_binary
        esac
    done
  )
}

function plugin::run_test() {
  go test -timeout=1m -bench=. -cover -v ./...
}

function plugin::build_binary() {
  go build -o "${ROOT}/go/bin/gpu-$arg" -ldflags "$(plugin::version::ldflags) -s -w" ${PACKAGE}/cmd/$arg
}

function plugin::generate_img() {
  readonly local commit=$(git log --no-merges --oneline | wc -l | sed -e 's,^[ \t]*,,')
  readonly local version=$(<"${ROOT}/VERSION")
  readonly local base_img=${BASE_IMG:-"thomassong/vcuda:1.0.4"}

  mkdir -p "${ROOT}/go/build"
  tar czf "${ROOT}/go/build/gpu-manager-source.tar.gz" --transform 's,^,/gpu-manager-'${version}'/,' $(plugin::source_targets)

  cp -R "${ROOT}/build/"* "${ROOT}/go/build/"

  (
    cd ${ROOT}/go/build
    docker build \
        --network=host \
        --build-arg version=${version} \
        --build-arg commit=${commit} \
        --build-arg base_img=${base_img} \
        -t "${IMAGE_FILE}:${version}" .
  )
}

function plugin::fmt() {
  local unfmt_files=()
  for file in $(plugin::fmt_targets); do
    if [[ -n $(gofmt -d -s $file 2>&1) ]]; then
      unfmt_files+=($file)
    fi
  done
  if [[ ${#unfmt_files[@]} -gt 0 ]]; then
    echo "need fmt ${unfmt_files[@]}"
    exit 1
  fi
}

plugin::build "$@"
