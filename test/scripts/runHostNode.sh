#！/bin/bash
## SPDX-License-Identifier: Apache-2.0
## Copyright Authors of Spider


set -o errexit
set -o nounset
set -o pipefail
set -x

CONTAINER_NAME=elf-hostnode
CONTAINER_HOST_MAP_PORT=20080


if [ "$1" == "on" ] ; then
  docker stop ${CONTAINER_NAME} &>/dev/null  || true
  docker rm ${CONTAINER_NAME} &>/dev/null || true


  DOCKER_MASTER_NODE_NAME="$2"
  APT_HTTP_PROXY=${3:-""}

  #================
  echo "generate kubeconf "
  docker cp ${DOCKER_MASTER_NODE_NAME}:/etc/kubernetes/admin.conf  /tmp/admin.conf
  MASTER_IP=$( docker exec ${DOCKER_MASTER_NODE_NAME} ip a show eth0 | grep -oP '(?<=inet\s)[0-9]+(\.[0-9]+){3}'  )
  [ -n "${MASTER_IP}" ] || { echo "error, failed to master ip" ; exit 1 ; }
  BRIDGE_ID=$( docker inspect ${DOCKER_MASTER_NODE_NAME} | jq .[0].NetworkSettings.Networks.kind.NetworkID  | tr -d '"' )
  [ -n "${BRIDGE_ID}" ] || { echo "error, failed to find bridge id for master " ; exit 1 ; }
  sed -i 's?'${DOCKER_MASTER_NODE_NAME}'?'${MASTER_IP}'?' /tmp/admin.conf

  docker run -d --network ${BRIDGE_ID} \
      --privileged \
      -p ${CONTAINER_HOST_MAP_PORT}:80/tcp \
      -v /tmp/admin.conf:/admin.conf  \
      --name ${CONTAINER_NAME} \
      ubuntu:24.10 sleep infinity
  docker exec ${CONTAINER_NAME} bash -c " export http_proxy=${APT_HTTP_PROXY} && apt-get update && apt-get install -y iproute2 curl iputils-ping "

elif [ "$1" == "runAgent" ] ; then
  SOURCE_IMAGE="$2"
  SOURCE_BIN_PATH="/usr/bin/agent"
  INSPECT_BIN_PATH="/usr/bin/inspect"

  echo "copy ${SOURCE_IMAGE}:$SOURCE_BIN_PATH to container ${CONTAINER_NAME} and run "

  docker create --name temp-container ${SOURCE_IMAGE}
  docker cp temp-container:${SOURCE_BIN_PATH} /tmp/$( basename $SOURCE_BIN_PATH )
  docker cp temp-container:${INSPECT_BIN_PATH} /tmp/$( basename ${INSPECT_BIN_PATH} )
  docker rm temp-container

  docker cp /tmp/$( basename $SOURCE_BIN_PATH ) ${CONTAINER_NAME}:${SOURCE_BIN_PATH}
  docker cp /tmp/$( basename $INSPECT_BIN_PATH ) ${CONTAINER_NAME}:${INSPECT_BIN_PATH}
  docker exec ${CONTAINER_NAME} bash -c " export KUBECONFIG=/admin.conf  && setsid ${SOURCE_BIN_PATH} "

elif [ "$1" == "runProxyServer" ] ; then
  SOURCE_IMAGE="$2"
  SOURCE_BIN_PATH="/usr/sbin/proxy_server"
  echo "copy ${SOURCE_IMAGE}:$SOURCE_BIN_PATH to container ${CONTAINER_NAME}, and run "

  docker create --name temp-container ${SOURCE_IMAGE}
  docker cp temp-container:${SOURCE_BIN_PATH} /tmp/$( basename $SOURCE_BIN_PATH )
  docker rm temp-container

  docker cp /tmp/$( basename $SOURCE_BIN_PATH ) ${CONTAINER_NAME}:${SOURCE_BIN_PATH}
  # map to CONTAINER_HOST_MAP_PORT
  docker exec ${CONTAINER_NAME} bash -c "setsid ${SOURCE_BIN_PATH} -port=80 "

elif [ "$1" == "off" ] ; then
    docker stop ${CONTAINER_NAME} &>/dev/null || true
    docker rm ${CONTAINER_NAME} &>/dev/null || true
    exit 0

else
   echo "unknow $1"
   exit 1
fi


