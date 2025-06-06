# LocalRedirect Policy

## Introduction

The LocalRedirect policy is inspired by similar functionality in the [cilium](https://github.com/cilium/cilium) project. Based on cGroup eBPF technology, it redirects requests to local services on the same node when a Pod accesses specified services.

It supports request redirection for the following objects:
* Applications within Pods
* Applications on cluster nodes

![Redirection Diagram](../images/localRedirect.png)

## Features

The current version supports the following features:

* [x] Frontend supports targeting services or custom VIPs and ports
* [x] Backend supports Pod label selectors
* [x] Supports configuring cluster-wide QoS limits

> Note: Between Balancing Policy and LocalRedirect Policy instances, the frontend does not support binding to the same service or defining the same virtual address, as this would cause resolution conflicts.

> When the frontend of a Balancing Policy or LocalRedirect Policy uses a custom IP address, if the IP address conflicts with a Kubernetes service ClusterIP, the policy's forwarding rules will take precedence for resolution.

## Use Cases

* Implementing transparent redirection for Node-local DNS
  To improve CoreDNS service capability and avoid DNS avalanche effects, the open-source community introduced Node-local DNS to implement DNS caching on each node.

  In traditional approaches, the DNS configuration of Pods is modified to point to the virtual address of the local Node-local DNS, binding this virtual address on the node. However, when the Node-local DNS fails or is upgraded, this approach cannot provide high-availability DNS service for applications.

  By introducing the redirection capability of the LocalRedirect policy, transparent and highly available service redirection can be provided for Node-local DNS without modifying Pod DNS configurations or introducing new virtual addresses. This supports resolving service access to the original CoreDNS service during local Node-local DNS failures or upgrades.

  Optionally, cluster-wide QoS limits can be configured. On each node, when the redirection count for a service reaches the QoS limit per second, redirection will not be implemented for that service within the current second, allowing service resolution to follow the normal process. This feature can be used to effectively set the per-second QoS limit for node-local proxies.

## Policy Examples

Below is a YAML example where the frontend points to a service name:

```yaml
apiVersion: balancing.elf.io/v1beta1
kind: LocalRedirectPolicy
metadata:
  name: test-service
spec:
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

## QoS Limits

1. Enabling

    Method 1: When installing balancing, set the helm parameter values.feature.redirectQosLimit=100

    Method 2: After installing balancing, you can set `kubectl set env deployment/balancing-agent -n elf REDIRECT_QOS_LIMIT=100`
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
