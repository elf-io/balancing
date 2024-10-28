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

FAIL_ACCOUNT=0
TEST_ACCOUNT=0


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
  echo '   curl -s 127.0.0.1:'${K8S_PROXY_SERVER_MAPPING_PORT}' -d "{\"BackendUrl\":\"'${LOCALVAR_URL}'\",\"Timeout\":5,\"ForwardType\":\"'${LOCALVAR_METHOD}'\", \"EchoData\":\"Hello\"}" | jq . '
  MSG=$( curl -s 127.0.0.1:${K8S_PROXY_SERVER_MAPPING_PORT} -d "{\"BackendUrl\":\"${LOCALVAR_URL}\",\"Timeout\":5,\"ForwardType\":\"${LOCALVAR_METHOD}\", \"EchoData\":\"Hello\"}"  ) \
     || { echo "failed to visit the proxy server on master node" ; exit 1 ; }
  echo "${MSG}" | jq .
  if echo "${MSG}" | jq '.BackendResponse' | grep -e "${LOCALVAR_EXPECT}" &>/dev/null ; then
        echo -e "--------------------${GREEN}result: pass${NC}-------------------"
  else
        echo -e "--------------------${RED}result: fail${NC}---------------------"
        (( FAIL_ACCOUNT = FAIL_ACCOUNT + 1 ))
  fi
  (( TEST_ACCOUNT = TEST_ACCOUNT + 1 ))
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
  echo '   curl -s 127.0.0.1:'${HOST_PROXY_SERVER_MAPPING_PORT}' -d "{\"BackendUrl\":\"'${LOCALVAR_URL}'\",\"Timeout\":5,\"ForwardType\":\"'${LOCALVAR_METHOD}'\", \"EchoData\":\"Hello\"}" | jq . '
  MSG=$( curl -s 127.0.0.1:${HOST_PROXY_SERVER_MAPPING_PORT} -d "{\"BackendUrl\":\"${LOCALVAR_URL}\",\"Timeout\":5,\"ForwardType\":\"${LOCALVAR_METHOD}\", \"EchoData\":\"Hello\"}"  ) \
     || { echo "failed to visit the proxy server on master node" ; exit 1 ; }
  echo "${MSG}" | jq .
  if echo "${MSG}" | jq '.BackendResponse' | grep -e "${LOCALVAR_EXPECT}" &>/dev/null ; then
        echo -e "--------------------${GREEN}result: pass${NC}-------------------"
  else
        echo -e "--------------------${RED}result: fail${NC}---------------------"
        (( FAIL_ACCOUNT = FAIL_ACCOUNT + 1 ))
  fi
  (( TEST_ACCOUNT = TEST_ACCOUNT + 1 ))
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

GetServicePort(){
    LOCAL_SERVICE_NAME="$1"
    LOCAL_SERVICE_NAMESPACE="$2"

    INFO=$( kubectl  get service backendserver-service-normal | sed '1d' | awk '{print $5}' | tr ',' '\n' )
    for LINE in $INFO ; do
       PROTOCOL=${LINE##*/}
       TMP=${LINE%%/*}
       PORT=${TMP%%:*}
       NODEPORT=${TMP##*:}
       echo "$PROTOCOL"
       echo "$PORT"
       echo "$NODEPORT"
    done
}

TestService(){
      echo "===================== test balancing: k8s service  ========================="
      NODE_IP_LIST=$(kubectl  get node -o wide | sed '1 d' | awk '{print $6}' )

      SERVICE_NAME=backendserver-service-external
      NORMAL_CLUSTER_IP=$( kubectl  get service ${SERVICE_NAME} | sed '1 d' | awk '{print $3}' )
      EXTERNAL_IP=$( kubectl  get service $SERVICE_NAME   | sed '1 d' | awk '{print $4}' )
      PORTINFO=$( kubectl  get service ${SERVICE_NAME} | sed '1d' | awk '{print $5}' | tr ',' '\n' )
      for LINE in $PORTINFO ; do
           PROTOCOL=${LINE##*/}
           TMP=${LINE%%/*}
           PORT=${TMP%%:*}
           NODEPORT=${TMP##*:}
           echo "service PORT: protocol=$PROTOCOL PORT=$PORT NODEPORT=$NODEPORT"
           if [ "$PROTOCOL" == "TCP" ] ; then
                METHOD="http"
                PORT_URL="http://${NORMAL_CLUSTER_IP}:${PORT}"
                EXTERNAL_URL="http://${EXTERNAL_IP}:${PORT}"
           else
                METHOD="udp"
                PORT_URL="${NORMAL_CLUSTER_IP}:${PORT}"
                EXTERNAL_URL="${EXTERNAL_IP}:${PORT}"
           fi
           # cluster ip + port
           VisitK8s "${PORT_URL}"  "${METHOD}"  \
                      "http: visit the cluster ip $NORMAL_CLUSTER_IP PORT=$PORT protocol=$PROTOCOL, to backend-server service"  "backendserver"
          #
           VisitK8s "${EXTERNAL_URL}"  "${METHOD}"  \
                      "http: visit the external ip $EXTERNAL_IP PORT=$PORT protocol=$PROTOCOL, to backend-server service"  "backendserver"
          # node port
          if [ -n "$NODEPORT" ] ; then
              for NODE_IP in ${NODE_IP_LIST} ; do
                   if [ "$PROTOCOL" == "TCP" ] ; then
                        URL="http://${NODE_IP}:${NODEPORT}"
                   else
                        URL="${NODE_IP}:${NODEPORT}"
                   fi
                  VisitK8s "${URL}"  "$METHOD"  \
                        "http: visit the nodeIp $NODE_IP NODEPORT=$NODEPORT protocol=$PROTOCOL , to backend-server service"  "backendserver"
              done
          fi
      done

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

      # DOTO: TEST on host client
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
              "http: visit the virtual address of backend-server balancing service"  "redirectserver"
    VisitK8s "${ADDRESS}:80"  "udp"  \
              "udp: visit the virtual address of backend-server balancing service"  "redirectserver"
    VisitHost "http://${ADDRESS}:80"  "http"  \
              "http: visit the virtual address of backend-server balancing service"  "redirectserver"
    VisitHost "${ADDRESS}:80"  "udp"  \
              "udp: visit the virtual address of backend-server balancing service"  "redirectserver"


    SERVICE_CLUSTER_IP=$( kubectl  get service backendserver-balancing-pod | sed '1 d' | awk '{print $3}' )
    VisitK8s "http://${SERVICE_CLUSTER_IP}:80"  "http"  \
              "http: visit the pod ip of backend-server balancing service"  "redirectserver"
    VisitK8s "${SERVICE_CLUSTER_IP}:80"  "udp"  \
              "udp: visit the pod ip of backend-server balancing service"  "redirectserver"


    SERVICE_CLUSTER_IP=$( kubectl  get service backendserver-balancing-hostport | sed '1 d' | awk '{print $3}' )
    VisitK8s "http://${SERVICE_CLUSTER_IP}:80"  "http"  \
              "http: visit the hostPort of backend-server balancing service"  "redirectserver"
    VisitK8s "${SERVICE_CLUSTER_IP}:80"  "udp"  \
              "udp: visit the hostPort of backend-server balancing service"  "redirectserver"
    VisitHost "http://${SERVICE_CLUSTER_IP}:80"  "http"  \
              "http: visit the hostPort of backend-server balancing service"  "redirectserver"
    VisitHost "${SERVICE_CLUSTER_IP}:80"  "udp"  \
              "udp: visit the hostPort of backend-server balancing service"  "redirectserver"


    SERVICE_CLUSTER_IP=$( kubectl  get service backendserver-balancing-nodeproxy | sed '1 d' | awk '{print $3}' )
    VisitK8s "http://${SERVICE_CLUSTER_IP}:80"  "http"  \
              "http: visit the nodeProxy of backend-server balancing service"  "redirectserver"
    VisitK8s "${SERVICE_CLUSTER_IP}:80"  "udp"  \
              "udp: visit the nodeProxy of backend-server balancing service"  "redirectserver"
    VisitHost "http://${SERVICE_CLUSTER_IP}:80"  "http"  \
              "http: visit the nodeProxy of backend-server balancing service"  "redirectserver"
    VisitHost "${SERVICE_CLUSTER_IP}:80"  "udp"  \
              "udp: visit the nodeProxy of backend-server balancing service"  "redirectserver"

    # DOTO: TEST on host client

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
    echo "unknown args: $@ "
    exit 1
fi

echo "================================================="
if ((FAIL_ACCOUNT==0)) ; then
  echo -e "${GREEN}${CURRENT_FILENAME}: ${TEST_ACCOUNT} tests pass ${NC}"
else
  echo -e "${GREEN}${CURRENT_FILENAME}: $((TEST_ACCOUNT-FAIL_ACCOUNT)) tests pass${NC}"
  echo -e "${RED}${CURRENT_FILENAME}: ${FAIL_ACCOUNT} tests failed ${NC}"
  exit 1
fi
