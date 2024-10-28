# Quick Start

## kubernetes installation

```shell
# get the host address of api server , 
# it should not be the clusterIP of kubernetes service, to make sure the balancing works when kube-proxy is not installed
~# kubectl cluster-info
Kubernetes control plane is running at https://192.168.0.10:6443

helm install -n elf balancing ./charts \
	--set feature.apiServerHost=192.168.0.10 \
	--set feature.apiServerPort=6443

```

## docker installation on host

```shell
  IMAGE_TAG=96abcfc96d2b33266bc62d76ee947e646267dd6e
  docker run -d --net=host \
      --privileged \
      -e "KUBECONFIG=/config" \
      -v /tmp/config:/config  \
      -v /sys/fs:/sys/fs:rw \
      -v /proc:/host/proc \
      ghcr.io/elf-io/balancing-agent:${IMAGE_TAG}
```

