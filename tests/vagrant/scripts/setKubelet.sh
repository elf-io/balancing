#!/bin/bash
# Copyright 2024 Authors of elf-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail
set -x

KUBELET_INTERFACE=${1:-"eth1"}

IPV4_ADDRESS=$( ip addr show ${KUBELET_INTERFACE} | grep "inet " | awk '{print $2}' | cut -d'/' -f1 )
IPV6_ADDRESS=$( ip addr show ${KUBELET_INTERFACE} | grep "inet6 " |  grep -v "scope link" | awk '{print $2}' | cut -d'/' -f1 )

echo "set up node on interface ${KUBELET_INTERFACE} with ipv4 ${IPV4_ADDRESS} and ipv6 ${IPV6_ADDRESS}"
echo 'KUBELET_EXTRA_ARGS="--node-ip='${IPV4_ADDRESS}','${IPV6_ADDRESS}'"' > /etc/default/kubelet
systemctl daemon-reload && systemctl restart kubelet
