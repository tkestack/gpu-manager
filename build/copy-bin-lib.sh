#!/bin/bash

set -o pipefail
set -o errexit
set -o nounset

FILE=${FILE:-"/etc/gpu-manager/volume.conf"}
LIB_FILES=$(jq -r .volume[1].components.libraries[] ${FILE})
BIN_FILES=$(jq -r .volume[1].components.binaries[] ${FILE})
readonly NV_DIR="/usr/local/nvidia"
readonly FIND_BASE=${FIND_BASE:-"/usr/local/host"}

function check_arch() {
  local readonly lib=$1
  if [[ $(objdump -f ${lib} | grep -o "elf64-x86-64") == "elf64-x86-64" ]]; then
    echo "64"
  else
    echo ""
  fi
}

function copy_lib() {
  for target in $(find ${FIND_BASE} -name "${1}*" | grep -v "stubs"); do
    if [[ $(objdump -p ${target} 2>/dev/null | grep -o "SONAME") == "SONAME" ]]; then
      copy_directory ${target} "${NV_DIR}/lib$(check_arch ${target})"
    fi
  done
}

function copy_bin() {
  for target in $(find ${FIND_BASE} -name "${1}"); do
    if [[ -L ${target} ]]; then
      echo "${target} is symlink"
      continue
    fi
    copy_directory ${target} "${NV_DIR}/bin/"
  done
}

function copy_directory() {
  local readonly lib=$1
  local readonly path=$2

  echo "copy ${lib} to ${path}"
  cp --preserve=mode,ownership -Pf "${lib}" "${path}"
}

rm -rf ${NV_DIR}
mkdir -p ${NV_DIR}/{bin,lib,lib64}

for file in ${LIB_FILES[@]}; do
  copy_lib ${file}
done

for file in ${BIN_FILES[@]}; do
  copy_bin ${file}
done

# fix libvdpau_nvidia.so
(
  cd ${NV_DIR}/lib
  rm -rf libvdpau_nvidia.so
  rel_path=$(readlink -f libvdpau_nvidia.so.1)
  ln -s $(basename ${rel_path}) libvdpau_nvidia.so
)

(
  cd ${NV_DIR}/lib64
  rm -rf libvdpau_nvidia.so
  rel_path=$(readlink -f libvdpau_nvidia.so.1)
  ln -s $(basename ${rel_path}) libvdpau_nvidia.so
)

# fix libnvidia-ml.so
(
  cd ${NV_DIR}/lib
  rm -rf libnvidia-ml.so
  rel_path=$(readlink -f libnvidia-ml.so.1)
  ln -s $(basename ${rel_path}) libnvidia-ml.so
)

(
  cd ${NV_DIR}/lib64
  rm -rf libnvidia-ml.so
  rel_path=$(readlink -f libnvidia-ml.so.1)
  ln -s $(basename ${rel_path}) libnvidia-ml.so
)
