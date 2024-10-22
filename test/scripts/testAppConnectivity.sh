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

which jq &>/dev/null || { echo "please install jq" ; exit 1 ; }

VisitService(){
  LOCALVAR_URL="${1}"
  LOCALVAR_METHOD="${2}"

  echo ""
  echo "visit the ${LOCALVAR_METHOD} server ${LOCALVAR_URL} "
  MSG=$( curl -s 127.0.0.1:20090 -d '{"BackendUrl":"'${LOCALVAR_URL}'","Timeout":5,"ForwardType":"'${LOCALVAR_METHOD}'", "EchoData":"Hello, HTTP!"}'  ) \
     || { echo "failed to visit the proxy server on master node" ; exit 1 ; }
  echo "${MSG}" | jq .

}

echo ""
echo "------------- test proxy-server by hostPort ------------ "
echo "visit the proxy server on master"
curl -s 127.0.0.1:20090/healthy || { echo "failed to visit the proxy server on master node" ; exit 1 ; }

echo ""
echo "visit the proxy server on worker"
curl -s 127.0.0.1:20091/healthy || { echo "failed to visit the proxy server on worker node" ; exit 1 ; }


echo ""
echo "------------- test backend-server  ------------ "
POD_NAMESPACE=default
POD_LABEL="app.kubernetes.io/instance=backendserver"
POD_IP_LIST=$( kubectl get pods --no-headers --kubeconfig ${E2E_KUBECONFIG}  --namespace ${POD_NAMESPACE} --selector ${POD_LABEL} --output jsonpath={.items[*].status.podIP} )
[ -z "${POD_IP_LIST}" ] && echo "error, failed to find the pod ip of backend server " && exit 1
for POD_IP in $POD_IP_LIST  ; do
      VisitService "http://${POD_IP}:80"  "http"
      VisitService "${POD_IP}:80"  "udp"
done

echo ""
echo "------------- test redirect-server  ------------ "
POD_NAMESPACE=default
POD_LABEL="app.kubernetes.io/instance=redirectserver"
POD_IP_LIST=$( kubectl get pods --no-headers --kubeconfig ${E2E_KUBECONFIG}  --namespace ${POD_NAMESPACE} --selector ${POD_LABEL} --output jsonpath={.items[*].status.podIP} )
[ -z "${POD_IP_LIST}" ] && echo "error, failed to find the pod ip of backend server " && exit 1
for POD_IP in $POD_IP_LIST  ; do
      VisitService "http://${POD_IP}:80"  "http"
      VisitService "${POD_IP}:80"  "udp"
done


echo "------------- test balancing: k8s service  ------------ "
NODE_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get node -o wide | sed -n '2 p' | awk '{print $6}' )

echo ""
echo "visit the cluster ip of normal service "
NORMAL_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-service-normal | sed '1 d' | awk '{print $3}' )
VisitService "http://${NORMAL_CLUSTER_IP}:80"  "http"
VisitService "${NORMAL_CLUSTER_IP}:80"  "udp"

echo ""
echo "visit the nodeport of normal service "
NODE_PORT_LIST=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-service-normal | sed '1d' | awk '{print $5}' | tr ',' '\n' | awk -F ':' '{print $2}' | grep -Eo "[0-9]+" )
NODE_PORT_IP=""
for PORT in ${NODE_PORT_LIST}; do
    ADDR="${NODE_IP}:${PORT}"
    VisitService "http://${ADDR}:80"  "http"
    VisitService "${ADDR}:80"  "udp"
done

echo ""
echo "visit the externalIp of normal service "
EXTERNAL_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-service-external   | sed '1 d' | awk '{print $4}' )
VisitService "http://${EXTERNAL_IP}:80"  "http"
VisitService "${EXTERNAL_IP}:80"  "udp"

echo ""
echo "visit the clusterIp of local service "
LOCAL_SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-service-local | sed '1 d' | awk '{print $3}' )
VisitService "http://${LOCAL_SERVICE_CLUSTER_IP}:80"  "http"
VisitService "${LOCAL_SERVICE_CLUSTER_IP}:80"  "udp"

echo ""
echo "visit the clusterIp of affinity service "
AFFINITY_SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-service-affinity | sed '1 d' | awk '{print $3}' )
VisitService "http://${AFFINITY_SERVICE_CLUSTER_IP}:80"  "http"
VisitService "${AFFINITY_SERVICE_CLUSTER_IP}:80"  "udp"

echo "------------- test balancing: localRedirect policy  ------------ "

echo ""
echo "visit the clusterIp of localRedirect service "
SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-redirect-service | sed '1 d' | awk '{print $3}' )
VisitService "http://${SERVICE_CLUSTER_IP}:80"  "http"
VisitService "${SERVICE_CLUSTER_IP}:80"  "udp"

echo ""
echo "visit the specified address of localRedirect service "
ADDRESS=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get LocalRedirectPolicy example-matchaddress | sed '1d' | awk '{print $2}' )
VisitService "http://${ADDRESS}:80"  "http"
VisitService "${ADDRESS}:80"  "udp"


echo "------------- test balancing: balancing policy  ------------ "

echo ""
echo "visit the specified address of balancing service "
ADDRESS=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get BalancingPolicy example-matchaddress | sed '1d' | awk '{print $2}' )
VisitService "http://${ADDRESS}:80"  "http"
VisitService "${ADDRESS}:80"  "udp"

echo ""
echo "visit the podEndpoint of balancing service "
SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-balancing-pod | sed '1 d' | awk '{print $3}' )
VisitService "http://${SERVICE_CLUSTER_IP}:80"  "http"
VisitService "${SERVICE_CLUSTER_IP}:80"  "udp"


echo ""
echo "visit the hostPort of balancing service "
SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-balancing-hostport | sed '1 d' | awk '{print $3}' )
VisitService "http://${SERVICE_CLUSTER_IP}:80"  "http"
VisitService "${SERVICE_CLUSTER_IP}:80"  "udp"

echo ""
echo "visit the nodeProxy of balancing service "
SERVICE_CLUSTER_IP=$( kubectl --kubeconfig ${E2E_KUBECONFIG} get service backendserver-balancing-nodeproxy | sed '1 d' | awk '{print $3}' )
VisitService "http://${SERVICE_CLUSTER_IP}:80"  "http"
VisitService "${SERVICE_CLUSTER_IP}:80"  "udp"
