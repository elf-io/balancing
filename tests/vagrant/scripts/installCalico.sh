#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -x

CURRENT_FILENAME=`basename $0`
CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)


echo "install calico"

VERSION=v3.28.2
#curl https://raw.githubusercontent.com/projectcalico/calico/${VERSION}/manifests/calico.yaml -O
sed -i 's?docker.io?docker.m.daocloud.io?'  ${CURRENT_DIR_PATH}/calico-${VERSION}.yaml
kubectl apply -f ${CURRENT_DIR_PATH}/calico-${VERSION}.yaml

PullImage(){
    IMAGE_LIST=$( cat ${CURRENT_DIR_PATH}/calico-${VERSION}.yaml | grep "image:" | sort | uniq )
    cat <<EOF > ${CURRENT_DIR_PATH}/pull-calico-image.sh
    for  IMAGE in ${IMAGE_LIST}; do
        echo "pull image: ${IMAGE}"
        podman pull ${IMAGE}
    done
EOF
    chmod +x ${CURRENT_DIR_PATH}/pull-calico-image.sh
    ${CURRENT_DIR_PATH}/pull-calico-image.sh
}
PullImage

kubectl set env daemonset -n kube-system calico-node CALICO_IPV4POOL_IPIP=Never
kubectl set env daemonset -n kube-system calico-node CALICO_IPV4POOL_VXLAN=CrossSubnet
kubectl set env daemonset -n kube-system calico-node CALICO_IPV6POOL_VXLAN=CrossSubnet
kubectl set env daemonset -n kube-system calico-node CALICO_IPV4POOL_NAT_OUTGOING=true
kubectl set env daemonset -n kube-system calico-node CALICO_IPV6POOL_NAT_OUTGOING=true
kubectl set env daemonset -n kube-system calico-node IP_AUTODETECTION_METHOD="can-reach=192.168.0.10"
kubectl set env daemonset -n kube-system calico-node IP6_AUTODETECTION_METHOD="can-reach=fd00::10"
kubectl set env daemonset -n kube-system calico-node FELIX_IPV6SUPPORT=true
#kubectl set env daemonset -n kube-system calico-node FELIX_IPTABLESBACKEND=NFT

kubectl set env daemonset -n kube-system calico-node CALICO_IPV6POOL_CIDR="fc01::/48"

# 切换 隧道模式
cat <<EOF | kubectl apply -f -
apiVersion: crd.projectcalico.org/v1
kind: IPPool
metadata:
  name: default-ipv4-ippool
spec:
  blockSize: 26
  cidr: 172.20.0.0/16
  ipipMode: Never
  natOutgoing: true
  nodeSelector: all()
  vxlanMode: CrossSubnet
EOF

cat <<EOF | kubectl apply -f -
apiVersion: crd.projectcalico.org/v1
kind: IPPool
metadata:
  name: default-ipv6-ippool
spec:
  blockSize: 122
  cidr: fc01::/48
  ipipMode: Never
  natOutgoing: true
  nodeSelector: all()
  vxlanMode: CrossSubnet
EOF


# set assign_ipv6
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: calico-config
  namespace: kube-system
data:
  calico_backend: bird
  cni_network_config: |-
    {
      "name": "k8s-pod-network",
      "cniVersion": "0.3.1",
      "plugins": [
        {
          "type": "calico",
          "log_level": "info",
          "log_file_path": "/var/log/calico/cni/cni.log",
          "datastore_type": "kubernetes",
          "nodename": "__KUBERNETES_NODE_NAME__",
          "mtu": __CNI_MTU__,
          "ipam": {
              "type": "calico-ipam",
              "assign_ipv4": "true",
              "assign_ipv6": "true"
          },
          "policy": {
              "type": "k8s"
          },
          "kubernetes": {
              "kubeconfig": "__KUBECONFIG_FILEPATH__"
          }
        },
        {
          "type": "portmap",
          "snat": true,
          "capabilities": {"portMappings": true}
        },
        {
          "type": "bandwidth",
          "capabilities": {"bandwidth": true}
        }
      ]
    }
  typha_service_name: none
  veth_mtu: "0"
EOF

kubectl rollout restart -n kube-system daemonset/calico-node
kubectl rollout restart -n kube-system deployment/calico-kube-controllers
