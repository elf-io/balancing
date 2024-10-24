#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

CURRENT_FILENAME=`basename $0`
CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)

ImageName="$1"
[ -n "$ImageName" ] || { echo "error, miss ImageName  " >&2 && exit 1 ; }

EXCLUDE_VM_LIST=${EXCLUDE_VM_LIST:-""}
SPECIFIED_VM_LIST=${SPECIFIED_VM_LIST:-""}

RAW_INFO=` vagrant status `
RUNNING_INFO=` grep " running " <<< "$RAW_INFO" | grep -v "host-alone" `
RUNNING_NUM=` echo "$RUNNING_INFO" | wc -l `
(( $RUNNING_NUM == 0 )) && echo "error, no running VMs" >&2 && exit 1
RUNNING_NAMES=` awk '{print $1}' <<< "$RUNNING_INFO" | tr  '\n' ' ' `
echo "available VMs: $RUNNING_NAMES "

FILE_NAME=$( tr './:-' '_' <<< "$ImageName" )
FILE_NAME="${FILE_NAME}.tar"

echo "docker save $ $ImageName "
rm -f /tmp/${FILE_NAME}
docker save -o /tmp/${FILE_NAME} $ImageName

if [ -n "${SPECIFIED_VM_LIST}" ] ; then
  RUNNING_NAMES="${SPECIFIED_VM_LIST}"
fi

for VM in $RUNNING_NAMES ; do
    [ -n "$EXCLUDE_VM_LIST" ] && grep " $VM " <<< " ${EXCLUDE_VM_LIST} " &>/dev/null && echo "ignore vm $VM" && continue
    echo "load $ImageName to VM $VM "
    if ${CURRENT_DIR_PATH}/ssh-exec $VM " podman images &>/dev/null " &>/dev/null ; then
        if ${CURRENT_DIR_PATH}/ssh-exec $VM " podman images  " | awk '{ printf("%s:%s\n",$1,$2) }' | grep "$ImageName" &>/dev/null ; then
            echo "image $ImageName exists in VM $VM, ignore copy "
            continue
        fi
        ${CURRENT_DIR_PATH}/cpToVM $VM /tmp/${FILE_NAME}   /tmp/${FILE_NAME}
        ${CURRENT_DIR_PATH}/ssh-exec $VM "podman load -i /tmp/${FILE_NAME} "
        ${CURRENT_DIR_PATH}/ssh-exec $VM "rm -f /tmp/${FILE_NAME} "
    elif ${CURRENT_DIR_PATH}/ssh-exec $VM " crictl images &>/dev/null " &>/dev/null ; then
        if ${CURRENT_DIR_PATH}/ssh-exec $VM " crictl images  " | awk '{ printf("%s:%s\n",$1,$2) }' | grep "$ImageName" &>/dev/null ; then
            echo "image $ImageName exists in VM $VM, ignore copy "
            continue
        fi
        ${CURRENT_DIR_PATH}/cpToVM $VM /tmp/${FILE_NAME}   /tmp/${FILE_NAME}
        ${CURRENT_DIR_PATH}/ssh-exec $VM "ctr -n k8s.io images import /tmp/${FILE_NAME} "
        ${CURRENT_DIR_PATH}/ssh-exec $VM "rm -f /tmp/${FILE_NAME} "
    else
        if ${CURRENT_DIR_PATH}/ssh-exec $VM " docker images  " | awk '{ printf("%s:%s\n",$1,$2) }' | grep "$ImageName" &>/dev/null ; then
            echo "image $ImageName exists in VM $VM, ignore copy "
            continue
        fi
        ${CURRENT_DIR_PATH}/cpToVM $VM /tmp/${FILE_NAME}   /tmp/${FILE_NAME}
        ${CURRENT_DIR_PATH}/ssh-exec $VM "docker load -i /tmp/${FILE_NAME} "
        ${CURRENT_DIR_PATH}/ssh-exec $VM "rm -f /tmp/${FILE_NAME} "
    fi
done
rm -f /tmp/${FILE_NAME}
echo "finish"
