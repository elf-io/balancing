# Quick Start

## Installation Requirements

1. **Linux Kernel Requirements for Host**: It is recommended that the Linux kernel version is greater than v5.8 to ensure all features function properly.

2. **Architecture Support**: Currently, only AMD architecture images are provided, and ARM architecture images are not yet available.

## Deploying Balancing in a Kubernetes Cluster

```shell
# Get the API Server access address
~# kubectl cluster-info
  Kubernetes control plane is running at https://192.168.0.10:6443

# Deploy Balancing, specifying the API Server access address separately, so that Balancing can access the API Server to complete its work even without running kube-proxy
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

## Deploying Balancing Agent Container Service on Hosts Outside the Cluster

You can deploy the Balancing Agent container service on hosts outside the cluster, and its configuration file should specify the configuration of the Kubernetes cluster to connect to.

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
