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

E2E_KUBECONFIG="$1"
[ -z "$E2E_KUBECONFIG" ] && echo "error, miss E2E_KUBECONFIG " && exit 1
[ ! -f "$E2E_KUBECONFIG" ] && echo "error, could not find file $E2E_KUBECONFIG " && exit 1
echo "$CURRENT_FILENAME : E2E_KUBECONFIG $E2E_KUBECONFIG "

echo ""
echo "------------- test proxy-server by hostPort ------------ "
echo "visit the proxy server on master"
curl -s 127.0.0.1:20090/healthy || { ehco "failed to visit the proxy server on master node" ; exit 1 ; }

echo ""
echo "visit the proxy server on worker"
curl -s 127.0.0.1:20091/healthy || { ehco "failed to visit the proxy server on worker node" ; exit 1 ; }


echo ""
echo "------------- test backend-server  ------------ "
POD_NAMESPACE=default
POD_LABEL="app.kubernetes.io/instance=backendserver"
POD_IP_LIST=$( kubectl get pods --no-headers --kubeconfig ${E2E_KUBECONFIG}  --namespace ${POD_NAMESPACE} --selector ${POD_LABEL} --output jsonpath={.items[*].status.podIP} )
[ -z "${POD_IP_LIST}" ] && echo "error, failed to find the pod ip of backend server " && exit 1
for POD_IP in $POD_IP_LIST  ; do
      echo ""
      echo "visit the http of backend server ${POD_IP} by proxy server"
      MSG=$( curl -s 127.0.0.1:20090 -d '{"BackendUrl":"http://'${POD_IP}':80","Timeout":5,"ForwardType":"http", "EchoData":"Hello, HTTP!"}'  ) \
         || { ehco "failed to visit the proxy server on master node" ; exit 1 ; }
      echo "${MSG}"

      echo ""
      echo "visit the udp of backend server ${POD_IP} by proxy server"
      MSG=$( curl -s 127.0.0.1:20090 -d '{"BackendUrl":"'${POD_IP}':80","Timeout":5,"ForwardType":"udp", "EchoData":"Hello, udp!"}'  ) \
         || { ehco "failed to visit the proxy server on master node" ; exit 1 ; }
      echo "${MSG}"
done

echo ""
echo "------------- test redirect-server  ------------ "
POD_NAMESPACE=default
POD_LABEL="app.kubernetes.io/instance=redirectserver"
POD_IP_LIST=$( kubectl get pods --no-headers --kubeconfig ${E2E_KUBECONFIG}  --namespace ${POD_NAMESPACE} --selector ${POD_LABEL} --output jsonpath={.items[*].status.podIP} )
[ -z "${POD_IP_LIST}" ] && echo "error, failed to find the pod ip of backend server " && exit 1
for POD_IP in $POD_IP_LIST  ; do
      echo ""
      echo "visit the http of redirect server ${POD_IP} by proxy server"
      MSG=$( curl -s 127.0.0.1:20090 -d '{"BackendUrl":"http://'${POD_IP}':80","Timeout":5,"ForwardType":"http", "EchoData":"Hello, HTTP!"}'  ) \
         || { ehco "failed to visit the proxy server on master node" ; exit 1 ; }
      echo "${MSG}"

      echo ""
      echo "visit the udp of redirect server ${POD_IP} by proxy server"
      MSG=$( curl -s 127.0.0.1:20090 -d '{"BackendUrl":"'${POD_IP}':80","Timeout":5,"ForwardType":"udp", "EchoData":"Hello, udp!"}'  ) \
         || { ehco "failed to visit the proxy server on master node" ; exit 1 ; }
      echo "${MSG}"
done

