#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

source copy-bin-lib.sh

echo "rebuild ldcache"
/usr/sbin/ldconfig

echo "launch gpu manager"
/usr/bin/gpu-manager --extra-config=/etc/gpu-manager/extra-config.json --v=${LOG_LEVEL} --hostname-override=${NODE_NAME} --share-mode=true --volume-config=/etc/gpu-manager/volume.conf --log-dir=/var/log/gpu-manager --query-addr=0.0.0.0 ${EXTRA_FLAGS:-""}