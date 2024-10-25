# Quick Start

## install on kubernetes

```shell
# get the host address of api server
~# kubectl cluster-info
Kubernetes control plane is running at https://192.168.0.10:6443

helm install -n elf balancing ./charts \
	--set feature.apiServerHost=192.168.0.10 \
	--set feature.apiServerPort=6443
 
```
## install on host

```shell
  docker run -d --net=host \
      --privileged \
      -v /tmp/config:/config  \
      -v /sys/fs:/sys/fs:rw \
      -v /proc:/host/proc \
      ghcr.io/elf-io/balancing-agent:v0.0.1 \
      bash -c "KUBECONFIG=/config /usr/bin/agent"
```

