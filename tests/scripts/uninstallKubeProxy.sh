#!/bin/bash
## SPDX-License-Identifier: Apache-2.0
## Copyright Authors of Spider

set -o errexit
set -o nounset
set -o pipefail
#set -x

[ -z "$KUBECONFIG" ] && echo "error, miss KUBECONFIG environment " && exit 1
[ ! -f "$KUBECONFIG" ] && echo "error, could not find file $KUBECONFIG " && exit 1
echo "KUBECONFIG ${KUBECONFIG} "

if ! kubectl get daemonset kube-proxy -n kube-system &>/dev/null ; then
    echo "warning: kube-proxy has been uninstalled"
    exit 0
fi

echo "clean kube-proxy rules"
kubectl patch daemonset kube-proxy -n kube-system --type='json' \
    -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/command", "value": ["/usr/local/bin/kube-proxy", "--cleanup=true"]}]'

echo "wait for pod"
for (( N =0 ; N<100; N++ )) ; do
    if kubectl get pod -n kube-system | grep kube-proxy | grep Completed &>/dev/null ; then
        break
    fi
    sleep 2
done

echo "wait for taking effect"
sleep 10

echo "delete daemonset "
kubeclt delete daemonset -n kube-system kube-proxy



