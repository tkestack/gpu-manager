#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

ROOT=$(cd $(dirname ${BASH_SOURCE[0]})/.. && pwd -P)

source "${ROOT}/hack/common.sh"

function plugin::build() {
  (
    plugin::setup_env
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

function plugin::test_package() {
  (
    cd "$GOPATH/src/$PACKAGE"
    subpackage=$(find . -not \( \( -wholename '*/vendor/*' \) -prune \) -name '*_test.go' -print|sed 's,./,,'|xargs -n1 dirname|sort -u)
    echo "$subpackage"
  )
}

function plugin::run_test() {
  local log=$(mktemp)
  echo "run test, log in $log"
  for p in $(plugin::test_package);do
    go test -timeout=1m -bench=. -cover -v "$PACKAGE/$p" 2>&1|tee -a $log || (cat $log && exit 1)
  done
  rm -rf ${log}
}

function plugin::build_binary() {
  go build -o "${GOPATH}/bin/gpu-$arg" -ldflags "$(plugin::version::ldflags) -s -w" ${PACKAGE}/cmd/$arg
}

function plugin::generate_img() {
  readonly local commit=$(git log --no-merges --oneline | wc -l | sed -e 's,^[ \t]*,,')
  readonly local version=$(<"${ROOT}/VERSION")
  readonly local base_img=${BASE_IMG:-"tkestack.io/public/vcuda:latest"}

  mkdir -p "${ROOT}/go/build"
  tar czf "${ROOT}/go/build/gpu-manager-source.tar.gz" --transform 's,^,/gpu-manager-'${version}'/,' $(plugin::source_targets)

  cp -R "${ROOT}/build/"* "${ROOT}/go/build/"

  (
    cd ${ROOT}/go/build
    docker build \
        --build-arg version=${version} \
        --build-arg commit=${commit} \
        --build-arg base_img=${base_img} \
        -t $IMAGE_FILE .
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
