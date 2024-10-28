#!/bin/bash
# Copyright 2024 Authors of elf-io
# SPDX-License-Identifier: Apache-2.0

CURRENT_FILENAME=`basename $0`
CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)

G_SERVICE_NAME="setIptablesSnat.service"
G_SERVICE_PATH="/usr/lib/systemd/system/${G_SERVICE_NAME}"
EXEC_PATH=${CURRENT_DIR_PATH}/setIptablesSnat.sh

addService(){

    #========================
    mkdir -p $(  dirname ${EXEC_PATH} )
    cat <<EOF > ${EXEC_PATH}
#!/bin/bash

# on or off
ACTION=\${1:-""}

set -x
set -o errexit
set -o nounset

sysctl -w net.ipv4.conf.all.forwarding=1 || true
sysctl -w net.ipv4.ip_forward=1 || true
sysctl -w net.ipv6.conf.all.forwarding=1 || true

removeIptablesByComment(){
   IPTABLES_CMD=\${1}
   IPTABLES_TABLE=\${2}
   IPTABLES_CHAIN=\${3}
   IPTABLES_COMMENT=\${4}

    while :; do
        if \${IPTABLES_CMD} -w -t \${IPTABLES_TABLE} -nxvL \${IPTABLES_CHAIN} 2>/dev/null | grep "\${IPTABLES_COMMENT}" &>/dev/null  ; then
            LINE=\$( \${IPTABLES_CMD} -w -t \${IPTABLES_TABLE} -nxvL \${IPTABLES_CHAIN} --line 2>/dev/null | grep "\${IPTABLES_COMMENT}" | awk '{print \$1}' | head -1 ) || true
            if grep -E '^[0-9]+$' <<< "\${LINE}" ; then
                echo "deleting line \${LINE} for table \${IPTABLES_TABLE}, chain \${IPTABLES_CHAIN} "
                \${IPTABLES_CMD} -w -t \${IPTABLES_TABLE} -D  \${IPTABLES_CHAIN}  \${LINE}
            fi
        else
            break
        fi
    done
}

if [ "\${ACTION}"x == "off"x ] ; then
    echo "all rules are off"
else
    echo "all rules are on"
fi

IPTABLES_COMMENT="set_snat_for_nodes"
removeIptablesByComment "iptables" "nat" "POSTROUTING" \${IPTABLES_COMMENT}
[ "\${ACTION}"x == "off"x ] \
  || iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE  --random-fully -m comment --comment \${IPTABLES_COMMENT}

IPTABLES_COMMENT="set_snat_for_nodes"
removeIptablesByComment "ip6tables" "nat" "POSTROUTING" \${IPTABLES_COMMENT}
[ "\${ACTION}"x == "off"x ] \
  || ip6tables -t nat -A POSTROUTING -o eth0 -j MASQUERADE  --random-fully -m comment --comment \${IPTABLES_COMMENT}


IPTABLES_COMMENT="accept_forward_for_nodes"
removeIptablesByComment "iptables" "filter" "FORWARD" \${IPTABLES_COMMENT}
[ "\${ACTION}"x == "off"x ] \
  || iptables -t filter -A FORWARD -j ACCEPT -m comment --comment \${IPTABLES_COMMENT}

IPTABLES_COMMENT="accept_forward_for_nodes"
removeIptablesByComment "ip6tables" "filter" "FORWARD" \${IPTABLES_COMMENT}
[ "\${ACTION}"x == "off"x ] \
  || ip6tables -t filter -A FORWARD -j ACCEPT -m comment --comment \${IPTABLES_COMMENT}

echo "finish"

EOF
    chmod 777 $EXEC_PATH

    rm -f $G_SERVICE_PATH
    cat << EOF > $G_SERVICE_PATH
[Unit]
Description=${G_SERVICE_NAME}
After=network.target

[Service]
Type=oneshot
ExecStart=$EXEC_PATH
ExecReload=$EXEC_PATH
Restart=no
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload &> /dev/null
    systemctl start $G_SERVICE_NAME 2> /dev/null
    systemctl enable $G_SERVICE_NAME 2> /dev/null
    systemctl status $G_SERVICE_NAME
}


rmService(){
    systemctl stop $G_SERVICE_NAME 2> /dev/null || true
    systemctl disable $G_SERVICE_NAME 2> /dev/null  || true
    [ -f "$G_SERVICE_PATH" ] && rm $G_SERVICE_PATH -f &> /dev/null
    systemctl daemon-reload &> /dev/null

    if [ -f "${EXEC_PATH}" ] ; then
        ${EXEC_PATH} "off" || true
        rm ${EXEC_PATH} -f &> /dev/null || true
    fi
}


if [ "${1}" == "off" ] ; then
    echo "set routerSnat off"
    rmService
else
    echo "set routerSnat on"
    addService
fi

iptables -t nat -nxvL POSTROUTING
ip6tables -t nat -nxvL POSTROUTING
iptables -t filter -nxvL FORWARD
ip6tables -t filter -nxvL FORWARD

