# Quick Start

## 安装要求

1. 主机的 Linux 内核要求。linux 内核最优大于 v5.8，所有功能都能够运行

2. 当前支持提供 AMD 架构镜像，未提供 ARM 架构镜像

## kubernetes 集群部署 Balancing

```shell
# 获取 api server 的访问地址
~# kubectl cluster-info
  Kubernetes control plane is running at https://192.168.0.10:6443

# 部署 balacning，其中单独指定了 api server 访问地址，这样，当不运行 kube-proxy 时， Balancing 依然能够访问 api server 完成工作
~# helm repo add elf https://elf-io.github.io/balancing
~# helm install -n elf balancing elf/balancing \
	--set feature.apiServerHost=192.168.0.10 \
	--set feature.apiServerPort=6443

~# kubectl get pod -n elf
  NAME                                    READY   STATUS    RESTARTS   AGE
  balancing-agent-jj8vq                   1/1     Running   0          4d10h
  balancing-agent-wqs4g                   1/1     Running   0          4d10h
  balancing-controller-849c9bd8f6-gbw6w   1/1     Running   0          4d10h
```

## 在集群外的主机上部署 Balancing agent 容器服务

可在集群外的主机上，部署 Balancing agent 容器服务，其配置文件中指明接入的 kubernetes 集群配置

```shell
  IMAGE_TAG=v0.0.2
  docker run -d --net=host \
      --privileged \
      -e "KUBECONFIG=/config" \
      -v ./config:/config  \
      -v /sys/fs:/sys/fs:rw \
      -v /proc:/host/proc \
      ghcr.io/elf-io/balancing-agent:${IMAGE_TAG}
```
