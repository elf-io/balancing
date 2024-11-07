# Balancing

Currently, Balancing is in the development and testing phase and is not suitable for production environments.

## Introduction to Balancing

Balancing is a layer 4 load balancing component implemented with eBPF on the Kubernetes platform. It references the functionalities of projects like [cilium](https://github.com/cilium/cilium), [calico](https://github.com/projectcalico/calico), and [KPNG](https://github.com/kubernetes-retired/kpng). Balancing supports running in a containerized manner within a Kubernetes cluster and also supports running in binary form on bare metal, providing CNI-independent load balancing access capabilities for applications inside and outside the Kubernetes cluster.

### Current Features

1. **CNI-independent cluster service resolution**
    - Implements service resolution initiated by Pods and Nodes on cluster nodes based on cGroup eBPF, achieving kube-proxy replacement.
    - Implements client-side load balancing resolution on external bare metal and virtual machines to support access to services within the Kubernetes cluster.
    - Future versions will support north-south nodePort resolution on node network cards based on TC eBPF.
    - For more information, please refer to [service resolution](./usages/service.en.md)

2. **Local redirection layer 4 load balancing resolution**:
    - Provides service redirection resolution initiated by Pods and Nodes based on cGroup eBPF to services on the same node, typical scenarios include local coreDns.
    - For more information, please refer to [LocalRedirect Policy](./usages/localredirect.en.md)

3. **Global and external service layer 4 load balancing resolution**:
    - Supports more flexible policy definitions, providing global load balancing strategies for applications inside and outside the cluster.
    - Application scenarios include client-side load balancing resolution for external hosts and load balancing within Kubernetes clusters.
    - For more information, please refer to [Balancing Policy](./usages/balancing.en.md)

4. **Event logging for resolution metrics**:
    - Records load balancing resolution events and associates related container information to form complete load balancing resolution metrics.

## Typical Use Cases

1. **Replacing kube-proxy service resolution in CNI-independent clusters**:
    - Suitable for underlay CNI that cannot implement service, such as [Macvlan](https://github.com/containernetworking/plugins/tree/main/plugins/main/macvlan), [SR-IOV CNI](https://github.com/k8snetworkplumbingwg/sriov-cni), etc.
    - Suitable for CNIs that do not implement eBPF functionality, such as [Antrea](https://github.com/antrea-io/antrea), [Kube-ovn](https://github.com/kubeovn/kube-ovn), [Flannel](https://github.com/flannel-io/flannel), and public cloud clusters.

2. **Service access redirection to local proxies**:
    - Implements high-availability redirection, directing coreDns services to Node-local DNS.
    - Implements node api-server proxy for [clusterpedia](https://github.com/clusterpedia-io/clusterpedia).

3. **Implementing eBPF layer 4 load balancing on the client side of external applications to access services in Kubernetes clusters**:
    - Applications in bare metal, virtual machines.
    - Applications in [kubevirt](https://github.com/kubevirt/kubevirt) virtual machines.
    - Edge nodes in [kubeedge](https://github.com/kubeedge/kubeedge) accessing cloud services (in progress).

4. **Layer 4 load balancing access between multiple clusters**:
    - Cross-cluster service access (in progress).

5. **Providing high-availability load balancing access entry for external bare metal services**:
    - Provides cluster internal load balancing access addresses for external applications through custom load balancing strategies and implements health checks (in progress).

## Architecture

![arch](./images/arch.png)

The Balancing component consists of an agent and a controller:
- **controller deployment**: Performs webhook validation and modification of various CRD objects.
- **agent daemonset**: Loads eBPF programs and distributes forwarding rules.

![eBPF](./images/cgroup-ebpf.png)

## Quick Start

- Refer to [Installation](./usages/install.en.md) for quick deployment.
- Refer to [Service Resolution](./usages/service.en.md) for usage experience.
- Refer to [LocalRedirect Policy](./usages/localredirect.en.md) for usage experience.
- Refer to [Balancing Policy](./usages/balancing.en.md) for usage experience.

## Roadmap

- **IP Family and Protocol**
  - [x] Support TCP and UDP
  - [x] Support IPv4
  - [ ] Support IPv6

- **Observability**
  - [x] Load balancing resolution logs
  - [ ] Load balancing resolution metrics

- **Service Resolution**
  - [x] East-west service resolution
  - [ ] North-south service resolution
  - [ ] sessionAffinity forwarding record health status

- **LocalRedirect Policy**
  - [x] Front supports pointing to service and custom VIP
  - [x] Backend supports pod label selectors

- **Balancing Policy**
  - [x] Front supports pointing to service and custom VIP
  - [x] Backend supports pod label selectors
  - [x] Backend supports custom IP and port
  - [ ] Inter-node forwarding tunnel

- **Multi-cluster Interconnection**
  - [ ] Cross-cluster service interconnection
  - [ ] Cross-cluster balancing policy

- **Others**
  - [x] Support amd architecture
  - [ ] Support arm architecture

## License

Balancing follows the Apache License, Version 2.0. For details, see [LICENSE](./LICENSE).
