# Local Redirect

## Introduction

The Local Redirect feature is designed to optimize network traffic within a Kubernetes cluster by redirecting traffic locally when possible. This feature is inspired by the capabilities of projects like [Cilium](https://github.com/cilium/cilium) and [Calico](https://github.com/projectcalico/calico).

## Features

Currently, it supports the following functionalities:

* [x] Local traffic redirection for Pods and Nodes: Ensures that traffic intended for services within the same node is redirected locally, reducing latency and improving performance.
* [x] Supports sessionAffinity based on ClientIP, ensuring consistent routing of client requests to the same backend Pod.

In future versions, the following enhancements are planned:

* [ ] Improved handling of external traffic: Ensuring that traffic from outside the cluster is efficiently redirected to the appropriate local services.
* [ ] Enhanced health checks for backend Pods to ensure traffic is only redirected to healthy instances.

## Use Cases

* By leveraging local redirection, clusters can achieve better performance and reduced network overhead, especially in environments with high inter-Pod communication.
* This feature is particularly beneficial for applications with high traffic volumes and latency-sensitive operations.

  > Note: Projects like [Cilium](https://github.com/cilium/cilium) and [Calico](https://github.com/projectcalico/calico) already include similar functionalities, making them suitable for clusters that require advanced network optimizations.

## Policy Examples

Below is a YAML example where the frontend points to a service name:

```yaml
apiVersion: balancing.elf.io/v1beta1
kind: LocalRedirectPolicy
metadata:
  name: test-service
spec:
  enabled: true
  frontend:
    serviceMatcher:
      serviceName: http-server-v4
      namespace: default
      toPorts:
        - port: "8080"
          protocol: TCP
          name: p1
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    endpointSelector:
      matchLabels:
        app: http-redirect
    toPorts:
      - port: "80"
        protocol: TCP
        name: p1
      - port: "80"
        protocol: TCP
        name: p2
```

Below is a YAML example where the frontend uses a custom virtual IP and port:

```yaml
apiVersion: balancing.elf.io/v1beta1
kind: LocalRedirectPolicy
metadata:
  name: test-addr
  annotations:
     balancing.elf.io/serviceId: "10091"
spec:
  enabled: true
  frontend:
    addressMatcher:
      ip: "169.254.0.90"
      toPorts:
        - port: "8080"
          protocol: TCP
          name: p1
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    endpointSelector:
      matchLabels:
        app: http-redirect
    toPorts:
      - port: "80"
        protocol: TCP
        name: p1
      - port: "80"
        protocol: TCP
        name: p2
```
