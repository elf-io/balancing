# Quick Start

## install on kubernetes

## install on host

```shell
  docker run -d --net=host \
      --privileged \
      -v /tmp/admin.conf:/admin.conf  \
      -v /sys/fs:/sys/fs \
      -v /proc:/host/proc \
      -v /var/run/balancing:/run/balancing \
      ghcr.io/elf-io/balancing-agent:v0.0.1 \
      bash -c "KUBECONFIG=/admin.conf /usr/bin/agent"

```

