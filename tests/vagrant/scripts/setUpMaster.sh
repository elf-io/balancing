#!/bin/bash
# Copyright 2024 Authors of elf-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail
set -x

CURRENT_FILENAME=`basename $0`
CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)


KUBELET_INTERFACE=${KUBELET_INTERFACE:-"eth1"}
POD_CIRD=${POD_CIRD:-"172.20.0.0/16,fc01::/48"}
LOCAL_CLUSTERIP_CIDR=${LOCAL_CLUSTERIP_CIDR:-"172.21.0.0/16,fc02::/108"}
WORKER_JOIN_SCRIPT_PATH=${WORKER_JOIN_SCRIPT_PATH:-"${CURRENT_DIR_PATH}/join.sh"}

SKIP_KUBE_PROXY=${SKIP_KUBE_PROXY:-""}

setUpMaster(){
    
    K8S_VERSION=$( kubeadm version | egrep -o "GitVersion[^,]+" | awk -F'"' '{print $2}' | tr -d '\n' )

    IPV4_ADDRESS=$( ip addr show ${KUBELET_INTERFACE} | grep "inet " | awk '{print $2}' | cut -d'/' -f1 )
    IPV6_ADDRESS=$( ip addr show ${KUBELET_INTERFACE} | grep "inet6 " |  grep -v "scope link" | awk '{print $2}' | cut -d'/' -f1 )
    echo "set up master on interface ${KUBELET_INTERFACE} with ipv4 ${IPV4_ADDRESS} and ipv6 ${IPV6_ADDRESS}"

    OPTION=""
    if [ "$SKIP_KUBE_PROXY" == "true" ] || [ "$SKIP_KUBE_PROXY" == "TRUE" ]  ; then
        echo "skip the installation of kube-proxy"
        OPTION+=" --skip-phases=addon/kube-proxy "
    else
        echo "install kube-proxy"
    fi

    kubeadm init  --v=5  --upload-certs --apiserver-advertise-address ${IPV4_ADDRESS} \
            --pod-network-cidr=${POD_CIRD}  --service-cidr ${LOCAL_CLUSTERIP_CIDR} \
            --token-ttl 0  --apiserver-cert-extra-sans 127.0.0.1 \
            --control-plane-endpoint ${IPV4_ADDRESS}:6443 \
            --kubernetes-version ${K8S_VERSION} ${OPTION}

    rm -rf /home/vagrant/.kube
    mkdir -p /home/vagrant/.kube
    cp -i /etc/kubernetes/admin.conf /home/vagrant/.kube/config
    sudo chown -R vagrant:vagrant /home/vagrant/.kube
    rm -rf /root/.kube
    mkdir -p /root/.kube
    cp -i /etc/kubernetes/admin.conf /root/.kube/config
    sudo chown -R root:root /root/.kube



    echo "wait for api-server ready"
    DONE="false"
    for (( n=0 ; n<= 60 ; n++)) ; do
        if kubectl get pod -A 2>/dev/null| grep apiserver | grep Running | grep "1/1" ; then
            echo "api server is ready" && DONE="true" && break
        fi
        echo "wait for api-server ready"
        sleep 10
    done
    [ "${DONE}" == "true" ] || { echo "error, master node is not ready " ; exit 1 ; }

    echo "generate join command"
    CLUSTER_TOKEN=$( kubeadm token list | sed '1 d'  | grep 'forever' | awk '{print $1}'  )
    CLUSTER_CERT=$( openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null | openssl dgst -sha256 -hex | awk '{print $2}' )
    [ -n "${CLUSTER_TOKEN}" ] || { echo "failed to get TOKEN from ${MASTER_HOSTNAME}" ; exit 1 ; }
    echo "cluster join TOKEN: ${CLUSTER_TOKEN}"
    [ -n "${CLUSTER_CERT}" ] || { echo "failed to get CERT from ${MASTER_HOSTNAME}" ; exit 1 ; }
    echo "cluster join CERT: ${CLUSTER_CERT}"
    MASTER_JOIN_KEY=$( kubeadm init phase upload-certs --upload-certs | sed -n '3 p' )
    [ -n "${MASTER_JOIN_KEY}" ] || { echo "failed to get certificate key from ${MASTER_HOSTNAME}" ; exit 1 ; }
    echo "master join key: ${MASTER_JOIN_KEY}"

    cat <<EOF > ${WORKER_JOIN_SCRIPT_PATH}
kubeadm join ${IPV4_ADDRESS}:6443 --token ${CLUSTER_TOKEN} --discovery-token-ca-cert-hash sha256:${CLUSTER_CERT} --v=5
EOF

    kubectl taint nodes --all node-role.kubernetes.io/master- &>/dev/null  || true
    kubectl taint nodes --all  node-role.kubernetes.io/control-plane- &>/dev/null || true

}

setUpMaster
