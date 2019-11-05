#!/usr/bin/env bash

readonly PACKAGE="tkestack.io/tkestack/gpu-manager"
readonly BUILD_IMAGE_REPO=plugin-build
readonly LOCAL_OUTPUT_IMAGE_STAGING="${ROOT}/go/images"
readonly IMAGE_FILE=${IMAGE_FILE:-"gpu-manager:latest"}
readonly PROTO_IMAGE="proto-generater"

function plugin::setup_env() {
  plugin::create_gopath_tree

  export GOPATH=${ROOT}/go
}

function plugin::create_gopath_tree() {
  local go_pkg_dir="${ROOT}/go/src/${PACKAGE}"
  local go_pkg_basedir=$(dirname "${go_pkg_dir}")

  mkdir -p "${go_pkg_basedir}"
  rm -f "${go_pkg_dir}"

  ln -s "${ROOT}" "${go_pkg_dir}"
}

function plugin::cleanup() {
  rm -rf ${ROOT}/go
}

function plugin::cleanup_image() {
  docker rm -vf ${PROTO_IMAGE}
}

function plugin::generate_proto() {
(
  PATH=${ROOT}/hack:$PATH
  protoc \
    --proto_path=vendor:. \
    --proto_path=staging/src:. \
    --proto_path=vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis:. \
    --go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:. \
    --grpc-gateway_out=logtostderr=true:. \
    pkg/api/runtime/display/api.proto

  protoc \
    --go_out=plugins=grpc:. \
    pkg/api/runtime/vcuda/api.proto
)
}

function plugin::version::ldflag() {
  local key=${1}
  local val=${2}
  echo "-X ${PACKAGE}/pkg/version.${key}=${val}"
}

function plugin::version::ldflags() {
  GIT_COMMIT=$(git log -1 --oneline 2>/dev/null | awk '{print $1}')
  local -a ldflags=()
  if [[ -n ${GIT_COMMIT} ]]; then
    ldflags+=($(plugin::version::ldflag "gitCommit" "${GIT_COMMIT}"))
  fi

  echo "${ldflags[*]-}"
}

function plugin::source_targets() {
  local targets=(
    $(find . -mindepth 1 -maxdepth 1 -not \(        \
        \( -path ./go \) -prune  \
      \))
  )
  echo "${targets[@]}"
}

function plugin::fmt_targets() {
  local targets=(
    $(find . -not \(  \
        \( -path ./go \
        -o -path ./vendor \
        \) -prune \
        \) \
        -name "*.go" \
        -print \
    )
  )
  echo "${targets[@]}"
}