#!/usr/bin/env bash

readonly PACKAGE="tkestack.io/gpu-manager"
readonly BUILD_IMAGE_REPO=plugin-build
readonly LOCAL_OUTPUT_IMAGE_STAGING="${ROOT}/go/images"
readonly IMAGE_FILE=${IMAGE_FILE:-"thomassong/gpu-manager"}
readonly PROTO_IMAGE="proto-generater"

function plugin::cleanup() {
  rm -rf ${ROOT}/go
}

function plugin::cleanup_image() {
  docker rm -vf ${PROTO_IMAGE}
}

function plugin::generate_proto() {
(
  docker run --rm \
    -v ${ROOT}/pkg/api:/tmp/pkg/api \
    -v ${ROOT}/staging/src:/tmp/staging/src \
    -u $(id -u) \
    devsu/grpc-gateway \
      bash -c "cd /tmp && protoc \\
        --proto_path=staging/src:. \\
        --proto_path=/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis:. \\
        --go_out=plugins=grpc:. \\
        --grpc-gateway_out=logtostderr=true:. \\
        pkg/api/runtime/display/api.proto"

  docker run --rm \
    -v ${ROOT}/pkg/api:/tmp/pkg/api \
    -u $(id -u) \
    devsu/grpc-gateway \
      bash -c "cd /tmp && protoc \\
        --go_out=plugins=grpc:. \\
        pkg/api/runtime/vcuda/api.proto"
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
