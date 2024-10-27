#!/bin/bash
## SPDX-License-Identifier: Apache-2.0
## Copyright Authors of Spider

set -o errexit
set -o nounset
set -o pipefail
#set -x

CURRENT_FILENAME=$( basename $0 )
CURRENT_DIR_PATH=$(cd $(dirname $0); pwd)
PROJECT_ROOT_PATH=$( cd ${CURRENT_DIR_PATH}/../.. && pwd )

[ -z "$KUBECONFIG" ] && echo "error, miss KUBECONFIG environment " && exit 1
[ ! -f "$KUBECONFIG" ] && echo "error, could not find file $KUBECONFIG " && exit 1

which jq &>/dev/null || { echo "please install jq" ; exit 1 ; }

K8S_PROXY_SERVER_MAPPING_PORT=${K8S_PROXY_SERVER_MAPPING_PORT:-"27001"}
HOST_PROXY_SERVER_MAPPING_PORT=${HOST_PROXY_SERVER_MAPPING_PORT:-"27002"}

echo "KUBECONFIG ${KUBECONFIG} "
echo "K8S_PROXY_SERVER_MAPPING_PORT ${K8S_PROXY_SERVER_MAPPING_PORT}"
echo "HOST_PROXY_SERVER_MAPPING_PORT ${HOST_PROXY_SERVER_MAPPING_PORT}"


# 定义颜色的 ANSI 转义码
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color，用于重置颜色

VisitK8s(){
  LOCALVAR_URL="${1}"
  LOCALVAR_METHOD="${2}"
  LOCALVAR_TITLE="${3}"
  LOCALVAR_EXPECT="${4}"

  echo ""
  echo "-------------- to K8S: ${LOCALVAR_TITLE} -----------------"
  echo "visit the ${LOCALVAR_METHOD} server ${LOCALVAR_URL} from k8s pod"
  echo '   curl -s 127.0.0.1:'${K8S_PROXY_SERVER_MAPPING_PORT}' -d "{\"BackendUrl\":\"'${LOCALVAR_URL}'\",\"Timeout\":5,\"ForwardType\":\"'${LOCALVAR_METHOD}'\", \"EchoData\":\"Hello!\"}" '
  MSG=$( curl -s 127.0.0.1:${K8S_PROXY_SERVER_MAPPING_PORT} -d "{\"BackendUrl\":\"${LOCALVAR_URL}\",\"Timeout\":5,\"ForwardType\":\"${LOCALVAR_METHOD}\", \"EchoData\":\"Hello!\"}"  ) \
     || { echo "failed to visit the proxy server on master node" ; exit 1 ; }
  echo "${MSG}" | jq .
  if echo "${MSG}" | jq '.BackendResponse' | grep -e "${LOCALVAR_EXPECT}" &>/dev/null ; then
        echo -e "--------------------${GREEN}result: pass${NC}-------------------"
  else
        echo -e "--------------------${RED}result: fail${NC}---------------------"
  fi
  echo ""
}

VisitHost(){
  LOCALVAR_URL="${1}"
  LOCALVAR_METHOD="${2}"
  LOCALVAR_TITLE="${3}"
  LOCALVAR_EXPECT="${4}"

  echo ""
  echo "-------------- to Host: ${LOCALVAR_TITLE} -----------------"
  echo "visit the ${LOCALVAR_METHOD} server ${LOCALVAR_URL} from k8s pod"
  echo '   curl -s 127.0.0.1:'${HOST_PROXY_SERVER_MAPPING_PORT}' -d "{\"BackendUrl\":\"'${LOCALVAR_URL}'\",\"Timeout\":5,\"ForwardType\":\"'${LOCALVAR_METHOD}'\", \"EchoData\":\"Hello!\"}" '
  MSG=$( curl -s 127.0.0.1:${HOST_PROXY_SERVER_MAPPING_PORT} -d "{\"BackendUrl\":\"${LOCALVAR_URL}\",\"Timeout\":5,\"ForwardType\":\"${LOCALVAR_METHOD}\", \"EchoData\":\"Hello!\"}"  ) \
     || { echo "failed to visit the proxy server on master node" ; exit 1 ; }
  echo "${MSG}" | jq .
  if echo "${MSG}" | jq '.BackendResponse' | grep -e "${LOCALVAR_EXPECT}" &>/dev/null ; then
        echo -e "--------------------${GREEN}result: pass${NC}-------------------"
  else
        echo -e "--------------------${RED}result: fail${NC}---------------------"
  fi
  echo ""
}

TestBasicConnectity(){
    echo ""
    echo "------------- directly test proxy-server by hostPort ------------ "
    echo "directly visit the proxy server on master"
    curl -s 127.0.0.1:${K8S_PROXY_SERVER_MAPPING_PORT}/healthy || { echo "failed to visit the proxy server on master node" ; exit 1 ; }

    echo ""
    echo "directly visit the proxy server on host"
    curl -s 127.0.0.1:${HOST_PROXY_SERVER_MAPPING_PORT}/healthy || { echo "failed to visit the proxy server on host " ; exit 1 ; }

    echo ""
    echo "------------- directly test pod id of backend-server by proxy-server ------------ "
    POD_NAMESPACE=default
    POD_LABEL="app.kubernetes.io/instance=backendserver"
    POD_IP_LIST=$( kubectl get pods --no-headers   --namespace ${POD_NAMESPACE} --selector ${POD_LABEL} --output jsonpath={.items[*].status.podIP} )
    [ -z "${POD_IP_LIST}" ] && echo "error, failed to find the pod ip of backend server " && exit 1
    for POD_IP in $POD_IP_LIST  ; do
          VisitK8s "http://${POD_IP}:80"  "http"  \
              "http: directly visit the pod ip ${POD_IP} of backend-server"  "backendserver"
          VisitK8s "${POD_IP}:80"  "udp" \
              "udp: directly visit the pod ip ${POD_IP} of backend-server"  "backendserver"
    done

    echo ""
    echo "------------- directly test redirect-server by proxy-server  ------------ "
    POD_NAMESPACE=default
    POD_LABEL="app.kubernetes.io/instance=redirectserver"
    POD_IP_LIST=$( kubectl get pods --no-headers   --namespace ${POD_NAMESPACE} --selector ${POD_LABEL} --output jsonpath={.items[*].status.podIP} )
    [ -z "${POD_IP_LIST}" ] && echo "error, failed to find the pod ip of backend server " && exit 1
    for POD_IP in $POD_IP_LIST  ; do
          VisitK8s "http://${POD_IP}:80"  "http"  \
              "http: directly visit the pod ip ${POD_IP} of redirect-server"  "redirectserver"
          VisitK8s "${POD_IP}:80"  "udp" \
              "udp: directly visit the pod ip ${POD_IP} of redirect-server"  "redirectserver"
    done
}

TestService(){
      echo "===================== test balancing: k8s service  ========================="
      NODE_IP_LIST=$(kubectl  get node -o wide | sed '1 d' | awk '{print $6}' )

      echo ""
      NORMAL_CLUSTER_IP=$( kubectl  get service backendserver-service-normal | sed '1 d' | awk '{print $3}' )
      VisitK8s "http://${NORMAL_CLUSTER_IP}:80"  "http"  \
                "http: visit the cluster ip of backend-server normal service"  "backendserver"
      VisitK8s "${NORMAL_CLUSTER_IP}:80"  "udp"  \
                "http: visit the cluster ip of backend-server normal service"  "backendserver"

      echo ""
      NODE_PORT_LIST=$( kubectl  get service backendserver-service-normal | sed '1d' | awk '{print $5}' | tr ',' '\n' | awk -F ':' '{print $2}' | grep -Eo "[0-9]+" )
      NODE_PORT_IP=""
      for PORT in ${NODE_PORT_LIST}; do
          for NODE_IP in ${NODE_IP_LIST} ; do
              ADDR="${NODE_IP}:${PORT}"
              VisitK8s "http://${ADDR}"  "http"  \
                        "http: visit the nodePort ip of backend-server normal service"  "backendserver"
              VisitK8s "${ADDR}"  "udp"  \
                        "http: visit the nodePort ip of backend-server normal service"  "backendserver"
          done
      done

      echo ""
      EXTERNAL_IP=$( kubectl  get service backendserver-service-external   | sed '1 d' | awk '{print $4}' )
      VisitK8s "http://${EXTERNAL_IP}:80"  "http"  \
                "http: visit the external ip of backend-server normal service"  "backendserver"
      VisitK8s "${EXTERNAL_IP}:80"  "udp"  \
                "http: visit the external ip of backend-server normal service"  "backendserver"

      echo ""
      LOCAL_SERVICE_CLUSTER_IP=$( kubectl  get service backendserver-service-local | sed '1 d' | awk '{print $3}' )
      VisitK8s "http://${LOCAL_SERVICE_CLUSTER_IP}:80"  "http"  \
                "http: visit the clusterIp of backend-server local service"  "backendserver.*workervm"
      VisitK8s "${LOCAL_SERVICE_CLUSTER_IP}:80"  "udp"  \
                "http: visit the clusterIp of backend-server local service"  "backendserver.*workervm"

      echo ""
      AFFINITY_SERVICE_CLUSTER_IP=$( kubectl  get service backendserver-service-affinity | sed '1 d' | awk '{print $3}' )
      VisitK8s "http://${AFFINITY_SERVICE_CLUSTER_IP}:80"  "http"  \
                "http: visit the clusterIp of backend-server affinity service"  "backendserver"
      VisitK8s "${AFFINITY_SERVICE_CLUSTER_IP}:80"  "udp"  \
                "http: visit the clusterIp of backend-server affinity service"  "backendserver"

}

TestRedirectPolicy(){
    echo "===================== test balancing: localRedirect policy  ========================="

    SERVICE_CLUSTER_IP=$( kubectl  get service backendserver-redirect-service | sed '1 d' | awk '{print $3}' )
    VisitK8s "http://${SERVICE_CLUSTER_IP}:80"  "http"  \
              "http: visit the clusterIp of backend-server localRedirect service"  "redirectserver.*workervm"
    VisitK8s "${SERVICE_CLUSTER_IP}:80"  "udp"  \
              "udp: visit the clusterIp of backend-server localRedirect service"  "redirectserver.*workervm"

    ADDRESS=$( kubectl  get LocalRedirectPolicy redirect-matchaddress | sed '1d' | awk '{print $2}' )
    VisitK8s "http://${ADDRESS}:80"  "http"  \
              "http: visit the virtual address of backend-server localRedirect service"  "redirectserver.*workervm"
    VisitK8s "${ADDRESS}:80"  "udp"  \
              "udp: visit the virtual address of backend-server localRedirect service"  "redirectserver.*workervm"

}

TestBalancingPolicy(){
    echo "===================== test balancing: balancing policy  ========================="

    ADDRESS=$( kubectl  get BalancingPolicy balancing-matchaddress | sed '1d' | awk '{print $2}' )
    VisitK8s "http://${ADDRESS}:80"  "http"  \
              "http: visit the virtual address of backend-server localRedirect service"  "redirectserver"
    VisitK8s "${ADDRESS}:80"  "udp"  \
              "udp: visit the virtual address of backend-server localRedirect service"  "redirectserver"
    VisitHost "http://${ADDRESS}:80"  "http"  \
              "http: visit the virtual address of backend-server localRedirect service"  "redirectserver"
    VisitHost "${ADDRESS}:80"  "udp"  \
              "udp: visit the virtual address of backend-server localRedirect service"  "redirectserver"


    SERVICE_CLUSTER_IP=$( kubectl  get service backendserver-balancing-pod | sed '1 d' | awk '{print $3}' )
    VisitK8s "http://${SERVICE_CLUSTER_IP}:80"  "http"  \
              "http: visit the pod ip of backend-server localRedirect service"  "redirectserver"
    VisitK8s "${SERVICE_CLUSTER_IP}:80"  "udp"  \
              "udp: visit the pod ip of backend-server localRedirect service"  "redirectserver"


    SERVICE_CLUSTER_IP=$( kubectl  get service backendserver-balancing-hostport | sed '1 d' | awk '{print $3}' )
    VisitK8s "http://${SERVICE_CLUSTER_IP}:80"  "http"  \
              "http: visit the hostPort of backend-server localRedirect service"  "redirectserver"
    VisitK8s "${SERVICE_CLUSTER_IP}:80"  "udp"  \
              "udp: visit the hostPort of backend-server localRedirect service"  "redirectserver"
    VisitHost "http://${SERVICE_CLUSTER_IP}:80"  "http"  \
              "http: visit the hostPort of backend-server localRedirect service"  "redirectserver"
    VisitHost "${SERVICE_CLUSTER_IP}:80"  "udp"  \
              "udp: visit the hostPort of backend-server localRedirect service"  "redirectserver"


    SERVICE_CLUSTER_IP=$( kubectl  get service backendserver-balancing-nodeproxy | sed '1 d' | awk '{print $3}' )
    VisitK8s "http://${SERVICE_CLUSTER_IP}:80"  "http"  \
              "http: visit the nodeProxy of backend-server localRedirect service"  "redirectserver"
    VisitK8s "${SERVICE_CLUSTER_IP}:80"  "udp"  \
              "udp: visit the nodeProxy of backend-server localRedirect service"  "redirectserver"
    VisitHost "http://${SERVICE_CLUSTER_IP}:80"  "http"  \
              "http: visit the nodeProxy of backend-server localRedirect service"  "redirectserver"
    VisitHost "${SERVICE_CLUSTER_IP}:80"  "udp"  \
              "udp: visit the nodeProxy of backend-server localRedirect service"  "redirectserver"

}

if [ "$1"x == "basic"x ]; then
    TestBasicConnectity
elif [ "$1"x == "service"x ]; then
    TestService
elif [ "$1"x == "balancing"x ]; then
    TestBalancingPolicy
elif [ "$1"x == "redirect"x ]; then
    TestRedirectPolicy
else
    echo "unknow args: $@ "
    exit 1
fi
