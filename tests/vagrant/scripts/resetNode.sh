#!/bin/bash
# Copyright 2024 Authors of elf-io
# SPDX-License-Identifier: Apache-2.0

CURRENT_FILENAME=`basename $0`
CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)


ResetNode(){
    echo "reset node"
    ansible-playbook /etc/ansible/playbooks/cleanup.yml  
    rm -rf /etc/ansible || true
    systemctl stop etcd  
    rm /var/lib/etcd -rf  
    kubeadm reset -f  

    swapoff -a && sed -i "s/.*swap.*/#&/" /etc/fstab

    rm /root/.kube -f
    systemctl enable kubelet  
    systemctl restart kubelet 
    systemctl enable crio
    systemctl restart crio

    rm /etc/sysctl.d/99-sysctl.conf
    rm /etc/sysctl.conf
    cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
net.ipv4.conf.all.forwarding = 1
net.ipv6.conf.all.disable_ipv6 = 0
net.ipv6.conf.default.disable_ipv6 = 0
net.ipv6.conf.lo.disable_ipv6 = 0
net.ipv6.conf.all.accept_dad = 0
net.ipv6.conf.default.accept_dad = 0
net.netfilter.nf_conntrack_max = 1048576
net.ipv4.ip_local_port_range = 1024 65535
net.ipv6.conf.all.forwarding = 1

net.ipv4.conf.all.arp_filter = 0
net.ipv4.conf.default.rp_filter = 0
net.ipv4.conf.all.arp_ignore = 1
net.ipv4.conf.default.arp_ignore = 1
net.ipv4.conf.all.arp_announce = 1
net.ipv4.conf.default.arp_announce = 1

net.ipv4.neigh.default.gc_thresh1 = 0
net.ipv4.neigh.default.gc_thresh2 = 512
net.ipv4.neigh.default.gc_thresh3 = 8192
net.ipv6.neigh.default.gc_thresh1 = 0
net.ipv6.neigh.default.gc_thresh2 = 512
net.ipv6.neigh.default.gc_thresh3 = 8192
EOF
  sysctl --system

}

ResetNode
