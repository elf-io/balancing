#！/bin/bash
## SPDX-License-Identifier: Apache-2.0
## Copyright Authors of Spider

set -o errexit
set -o nounset
set -o pipefail
#set -x


CURRENT_FILENAME=$( basename $0 )
CURRENT_DIR_PATH=$(cd $(dirname $0); pwd)
PROJECT_ROOT_PATH=$( cd ${CURRENT_DIR_PATH}/../.. && pwd )

E2E_KUBECONFIG="${1}"
[ -z "$E2E_KUBECONFIG" ] && echo "error, miss E2E_KUBECONFIG " && exit 1
[ ! -f "$E2E_KUBECONFIG" ] && echo "error, could not find file $E2E_KUBECONFIG " && exit 1
echo "$CURRENT_FILENAME : E2E_KUBECONFIG $E2E_KUBECONFIG "
export KUBECONFIG=${E2E_KUBECONFIG}

which jq &>/dev/null || { echo "please install jq" ; exit 1 ; }

K8S_PROXY_SERVER_MAPPING_PORT="${2}"
HOST_PROXY_SERVER_MAPPING_PORT="${3}"
echo "K8S_PROXY_SERVER_MAPPING_PORT ${K8S_PROXY_SERVER_MAPPING_PORT}"
echo "HOST_PROXY_SERVER_MAPPING_PORT ${HOST_PROXY_SERVER_MAPPING_PORT}"


VisitServiceForK8s(){
  LOCALVAR_URL="${1}"
  LOCALVAR_METHOD="${2}"

  echo ""
  echo "visit the ${LOCALVAR_METHOD} server ${LOCALVAR_URL} from k8s pod"
  MSG=$( curl -s 127.0.0.1:${K8S_PROXY_SERVER_MAPPING_PORT} -d '{"BackendUrl":"'${LOCALVAR_URL}'","Timeout":5,"ForwardType":"'${LOCALVAR_METHOD}'", "EchoData":"Hello, HTTP!"}'  ) \
     || { echo "failed to visit the proxy server on master node" ; exit 1 ; }
  echo "${MSG}" | jq .

}

VisitServiceForAll(){
  LOCALVAR_URL="${1}"
  LOCALVAR_METHOD="${2}"

  echo ""
  echo "visit the ${LOCALVAR_METHOD} server ${LOCALVAR_URL} from k8s pod"
  MSG=$( curl -s 127.0.0.1:${K8S_PROXY_SERVER_MAPPING_PORT} -d '{"BackendUrl":"'${LOCALVAR_URL}'","Timeout":5,"ForwardType":"'${LOCALVAR_METHOD}'", "EchoData":"Hello, HTTP!"}'  ) \
     || { echo "failed to visit the proxy server on master node" ; exit 1 ; }
  echo "${MSG}" | jq .

  echo ""
  echo "visit the ${LOCALVAR_METHOD} server ${LOCALVAR_URL} from host"
  MSG=$( curl -s 127.0.0.1:${HOST_PROXY_SERVER_MAPPING_PORT} -d '{"BackendUrl":"'${LOCALVAR_URL}'","Timeout":5,"ForwardType":"'${LOCALVAR_METHOD}'", "EchoData":"Hello, HTTP!"}'  ) \
     || { echo "failed to visit the proxy server on master node" ; exit 1 ; }
  echo "${MSG}" | jq .

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
    echo "------------- directly test backend-server  ------------ "
    POD_NAMESPACE=default
    POD_LABEL="app.kubernetes.io/instance=backendserver"
    POD_IP_LIST=$( kubectl get pods --no-headers --kubeconfig ${E2E_KUBECONFIG}  --namespace ${POD_NAMESPACE} --selector ${POD_LABEL} --output jsonpath={.items[*].status.podIP} )
    [ -z "${POD_IP_LIST}" ] && echo "error, failed to find the pod ip of backend server " && exit 1
    for POD_IP in $POD_IP_LIST  ; do
          echo "directly visit the backend-server  "
          VisitServiceForK8s "http://${POD_IP}:80"  "http"
          VisitServiceForK8s "${POD_IP}:80"  "udp"
    done

    echo ""
    echo "------------- directly test redirect-server  ------------ "
    POD_NAMESPACE=default
    POD_LABEL="app.kubernetes.io/instance=redirectserver"
    POD_IP_LIST=$( kubectl get pods --no-headers --kubeconfig ${E2E_KUBECONFIG}  --namespace ${POD_NAMESPACE} --selector ${POD_LABEL} --output jsonpath={.items[*].status.podIP} )
    [ -z "${POD_IP_LIST}" ] && echo "error, failed to find the pod ip of backend server " && exit 1
    for POD_IP in $POD_IP_LIST  ; do
          echo "directly visit the redirect-server  "
          VisitServiceForK8s "http://${POD_IP}:80"  "http"
          VisitServiceForK8s "${POD_IP}:80"  "udp"
    done
}

TestService(){
      echo "===================== test balancing: k8s service  ========================="
      NODE_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get node -o wide | sed -n '2 p' | awk '{print $6}' )

      echo ""
      echo "----------- test balancing of k8S service: visit the cluster ip of backend-server normal service "
      NORMAL_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-service-normal | sed '1 d' | awk '{print $3}' )
      VisitServiceForAll "http://${NORMAL_CLUSTER_IP}:80"  "http"
      VisitServiceForAll "${NORMAL_CLUSTER_IP}:80"  "udp"

      echo ""
      echo "----------- test balancing of k8S service: visit the nodePort of backend-server normal service "
      NODE_PORT_LIST=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-service-normal | sed '1d' | awk '{print $5}' | tr ',' '\n' | awk -F ':' '{print $2}' | grep -Eo "[0-9]+" )
      NODE_PORT_IP=""
      for PORT in ${NODE_PORT_LIST}; do
          ADDR="${NODE_IP}:${PORT}"
          VisitServiceForAll "http://${ADDR}:80"  "http"
          VisitServiceForAll "${ADDR}:80"  "udp"
      done

      echo ""
      echo "----------- test balancing of k8S service: visit the externalIp of backend-server normal service "
      EXTERNAL_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-service-external   | sed '1 d' | awk '{print $4}' )
      VisitServiceForAll "http://${EXTERNAL_IP}:80"  "http"
      VisitServiceForAll "${EXTERNAL_IP}:80"  "udp"

      echo ""
      echo "----------- test balancing of k8S service: visit the clusterIp of backend-server local service "
      LOCAL_SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-service-local | sed '1 d' | awk '{print $3}' )
      VisitServiceForAll "http://${LOCAL_SERVICE_CLUSTER_IP}:80"  "http"
      VisitServiceForAll "${LOCAL_SERVICE_CLUSTER_IP}:80"  "udp"

      echo ""
      echo "----------- test balancing of k8S service: visit the clusterIp of backend-server affinity service "
      AFFINITY_SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-service-affinity | sed '1 d' | awk '{print $3}' )
      VisitServiceForAll "http://${AFFINITY_SERVICE_CLUSTER_IP}:80"  "http"
      VisitServiceForAll "${AFFINITY_SERVICE_CLUSTER_IP}:80"  "udp"
}

TestRedirectPolicy(){
    echo "===================== test balancing: localRedirect policy  ========================="

    echo ""
    echo "----------- test balancing of localRedirect policy: visit the clusterIp of backend-server localRedirect service "
    SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-redirect-service | sed '1 d' | awk '{print $3}' )
    VisitServiceForAll "http://${SERVICE_CLUSTER_IP}:80"  "http"
    VisitServiceForAll "${SERVICE_CLUSTER_IP}:80"  "udp"

    echo ""
    echo "----------- test balancing of localRedirect policy: visit the virtual address of backend-server localRedirect service "
    ADDRESS=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get LocalRedirectPolicy example-matchaddress | sed '1d' | awk '{print $2}' )
    VisitServiceForAll "http://${ADDRESS}:80"  "http"
VisitServiceForAll "${ADDRESS}:80"  "udp"
}

TestBalancingPolicy(){
    echo "===================== test balancing: balancing policy  ========================="

    echo ""
    echo "----------- test balancing of balancing policy: visit the virtual address of backend-server balancing service "
    ADDRESS=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get BalancingPolicy example-matchaddress | sed '1d' | awk '{print $2}' )
    VisitServiceForAll "http://${ADDRESS}:80"  "http"
    VisitServiceForAll "${ADDRESS}:80"  "udp"

    echo ""
    echo "----------- test balancing of balancing policy: visit the podEndpoint of backend-server balancing service "
    SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-balancing-pod | sed '1 d' | awk '{print $3}' )
    VisitServiceForAll "http://${SERVICE_CLUSTER_IP}:80"  "http"
    VisitServiceForAll "${SERVICE_CLUSTER_IP}:80"  "udp"


    echo ""
    echo "----------- test balancing of balancing policy: visit the hostPort of backend-server balancing service "
    SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-balancing-hostport | sed '1 d' | awk '{print $3}' )
    VisitServiceForAll "http://${SERVICE_CLUSTER_IP}:80"  "http"
    VisitServiceForAll "${SERVICE_CLUSTER_IP}:80"  "udp"

    echo ""
    echo "----------- test balancing of balancing policy: visit the nodeProxy of backend-server balancing service "
    SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-balancing-nodeproxy | sed '1 d' | awk '{print $3}' )
    VisitServiceForAll "http://${SERVICE_CLUSTER_IP}:80"  "http"
    VisitServiceForAll "${SERVICE_CLUSTER_IP}:80"  "udp"
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