#!/bin/bash

CURRENT_FILENAME=`basename $0`
CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)


# 定义镜像变量
K8S_IMAGE=${K8S_IMAGE:-"alvistack/kubernetes-1.30"}
UBUNTU_IMAGE=${UBUNTU_IMAGE:-"alvistack/ubuntu-24.04"}

# 定义资源变量
VM_MEMORY=${VM_MEMORY:-$((${VM_MEMORY:-1024}*8))}
VM_CPUS=${VM_CPUS:-"4"} 
HOSTPORT_API_SERVER=${HOSTPORT_API_SERVER:-"26443"}
KUBECONFIG_PATH=${KUBECONFIG_PATH:-${CURRENT_DIR_PATH}/config}

# 检查命令行参数
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 {on|off}"
  exit 1
fi

# 初始化 Vagrantfile
create_vagrantfile() {
  cat <<EOF > Vagrantfile
Vagrant.configure("2") do |config|

  # 定义 Kubernetes 主节点虚拟机
  config.vm.define "k8s-master" do |k8s|
    k8s.vm.box = "$K8S_IMAGE"
    k8s.vm.hostname = "k8s-master"
    k8s.vm.network "private_network", ip: "192.168.0.10", netmask: "255.255.255.0", ipv6: "fd00::10", ipv6_prefix_length: 64
    k8s.vm.network "forwarded_port", guest: 6443, host: ${HOSTPORT_API_SERVER}
    k8s.vm.network "forwarded_port", guest: 27000, host: 27000
    k8s.vm.provider "virtualbox" do |vb|
      vb.memory = "$VM_MEMORY"
      vb.cpus = "$VM_CPUS"
    end
    # 恢复默认的 vagrant 用户 SSH 登录
    # 挂载 scripts 目录
    k8s.vm.synced_folder "./scripts", "/home/vagrant/scripts"
    # 在虚拟机中运行 setUpMaster.sh 脚本
    k8s.vm.provision "shell", inline: <<-SHELL
      set -o errexit
      set -o nounset
      set -o pipefail

      # 确保 vagrant 用户具有 sudo 权限
      echo "vagrant ALL=(ALL) NOPASSWD:ALL" | sudo tee /etc/sudoers.d/vagrant
      sudo chmod 0440 /etc/sudoers.d/vagrant

      # the image disable ipv6 by default, so reconfigure it 
      sudo sysctl -w net.ipv6.conf.all.disable_ipv6=0
      sudo ip a a 192.168.0.10/24 dev eth1 || true
      sudo ip a a fd00::10/64 dev eth1 || true  

      # 生成 SSH 密钥对
      ssh-keygen -t rsa -b 2048 -f /home/vagrant/.ssh/id_rsa -N ""
      cat /home/vagrant/.ssh/id_rsa.pub >> /home/vagrant/.ssh/authorized_keys

      # 在 .bashrc 中添加 sudo -i
      echo "sudo -i" >> /home/vagrant/.bashrc

      # set up kubernetes master
      chmod +x /home/vagrant/scripts/resetNode.sh
      /home/vagrant/scripts/resetNode.sh 
      chmod +x /home/vagrant/scripts/getImages.sh
      /home/vagrant/scripts/getImages.sh
      chmod +x /home/vagrant/scripts/setUpMaster.sh
      export WORKER_JOIN_SCRIPT_PATH=/home/vagrant/scripts/join.sh
      sudo /home/vagrant/scripts/setUpMaster.sh
      chmod +x /home/vagrant/scripts/installCalico.sh
      /home/vagrant/scripts/installCalico.sh

      # 删除原有默认路由
      ip route del default || true
      ip -6 route del default || true
      # 设置新的默认路由
      ip route add default via 192.168.0.2
      ip -6 route add default via fd00::2

    SHELL
  end

  # 定义 Kubernetes 工作节点虚拟机
  config.vm.define "k8s-worker" do |k8s|
    k8s.vm.box = "$K8S_IMAGE"
    k8s.vm.hostname = "k8s-worker"
    k8s.vm.network "private_network", ip: "192.168.0.11", netmask: "255.255.255.0", ipv6: "fd00::11", ipv6_prefix_length: 64
    k8s.vm.provider "virtualbox" do |vb|
      vb.memory = "$VM_MEMORY"
      vb.cpus = "$VM_CPUS"
    end
    # 恢复默认的 vagrant 用户 SSH 登录
    # 挂载 scripts 目录
    k8s.vm.synced_folder "./scripts", "/home/vagrant/scripts"
    # 在虚拟机中运行 setUpWorker.sh 脚本
    k8s.vm.provision "shell", inline: <<-SHELL
      set -o errexit
      set -o nounset
      set -o pipefail

      # 确保 vagrant 用户具有 sudo 权限
      echo "vagrant ALL=(ALL) NOPASSWD:ALL" | sudo tee /etc/sudoers.d/vagrant
      sudo chmod 0440 /etc/sudoers.d/vagrant

      # the image disable ipv6 by default, so reconfigure it 
      sudo sysctl -w net.ipv6.conf.all.disable_ipv6=0
      sudo ip a a 192.168.0.11/24 dev eth1 || true
      sudo ip a a fd00::11/64 dev eth1 || true

      # 生成 SSH 密钥对
      ssh-keygen -t rsa -b 2048 -f /home/vagrant/.ssh/id_rsa -N ""
      cat /home/vagrant/.ssh/id_rsa.pub >> /home/vagrant/.ssh/authorized_keys

      # 在 .bashrc 中添加 sudo -i
      echo "sudo -i" >> /home/vagrant/.bashrc

      # set up kubernetes worker
      chmod +x /home/vagrant/scripts/resetNode.sh
      /home/vagrant/scripts/resetNode.sh 
      chmod +x /home/vagrant/scripts/getImages.sh
      /home/vagrant/scripts/getImages.sh 
      scp -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null 192.168.0.10:/home/vagrant/scripts/join.sh /home/vagrant/scripts/join.sh
      chmod +x /home/vagrant/scripts/join.sh
      sudo /home/vagrant/scripts/join.sh 

      # 删除原有默认路由
      ip route del default || true
      ip -6 route del default || true
      # 设置新的默认路由
      ip route add default via 192.168.0.2
      ip -6 route add default via fd00::2

    SHELL
  end

  # 定义 Ubuntu 虚拟机
  config.vm.define "host-alone" do |ubuntu|
    ubuntu.vm.box = "$UBUNTU_IMAGE"
    ubuntu.vm.hostname = "ubuntu-vm"
    ubuntu.vm.network "private_network", ip: "192.168.0.2", netmask: "255.255.255.0", ipv6: "fd00::2", ipv6_prefix_length: 64
    ubuntu.vm.network "forwarded_port", guest: 80, host: 27001
    ubuntu.vm.provider "virtualbox" do |vb|
      vb.memory = "$VM_MEMORY"
      vb.cpus = "$VM_CPUS"
    end
    # 恢复默认的 vagrant 用户 SSH 登录
    # 挂载 scripts 目录
    ubuntu.vm.synced_folder "./scripts", "/home/vagrant/scripts"
    # 启用 IP 转发和调整 iptables 规则
    ubuntu.vm.provision "shell", inline: <<-SHELL
      set -o errexit
      set -o nounset
      set -o pipefail

      # 确保 vagrant 用户具有 sudo 权限
      echo "vagrant ALL=(ALL) NOPASSWD:ALL" | sudo tee /etc/sudoers.d/vagrant
      sudo chmod 0440 /etc/sudoers.d/vagrant

      # the image disable ipv6 by default, so reconfigure it 
      sudo sysctl -w net.ipv6.conf.all.disable_ipv6=0
      sudo ip a a 192.168.0.2/24 dev eth1 || true
      sudo ip a a fd00::2/64 dev eth1 || true

      # 在 .bashrc 中添加 sudo -i
      echo "sudo -i" >> /home/vagrant/.bashrc

      # 启用 IPv4 和 IPv6 转发
      sysctl -w net.ipv4.ip_forward=1
      sysctl -w net.ipv6.conf.all.forwarding=1

      # 清除 iptables 规则
      iptables -F
      iptables -t nat -F
      iptables -t mangle -F
      iptables -X

      ip6tables -F
      ip6tables -t nat -F
      ip6tables -t mangle -F
      ip6tables -X

      # 设置 iptables 规则允许转发
      iptables -P FORWARD ACCEPT
      ip6tables -P FORWARD ACCEPT

    SHELL
  end
end
EOF
}

SetKubeconfig(){
    sshpass -p vagrant scp -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -P 2222 vagrant@127.0.0.1:/home/vagrant/.kube/config ${KUBECONFIG_PATH}
    sed -i -E 's?server: .*?server: https://127.0.0.1:'${HOSTPORT_API_SERVER}'?' ${KUBECONFIG_PATH}

    echo "wait for cluster ready"
    DONE=""
    for i in {1..60}; do
      echo "waiting for cluster ready: ${i}"
      KUBECONFIG=${KUBECONFIG_PATH} kubectl get pod -A | sed '1d' | grep -v Running &>/dev/null
      if [ $? -eq 0 ]; then
        sleep 10
        DONE="true"
        continue
      fi
      break
    done
    if [ -z "$DONE" ] ; then
       echo "k8s cluster is not ready, time out"
       exit 1
    fi

    echo "============================"
    KUBECONFIG=${KUBECONFIG_PATH} kubectl get pod -A
    echo "============================"

    echo "export KUBECONFIG=${KUBECONFIG_PATH}"
    echo ""
}


# 根据参数执行相应操作
case "$1" in
  on)
    echo "start time: $(date)"
    create_vagrantfile
    vagrant up
    SetKubeconfig
    echo "end time: $(date)"
    ;;
  off)
    vagrant destroy -f k8s-master k8s-worker host-alone
    ;;
  *)
    echo "Invalid command. Use 'on' to start the VMs or 'off' to destroy them."
    exit 1
    ;;
esac
