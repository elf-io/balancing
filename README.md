# Balancing

[**简体中文**](./docs/readme.zh.md)

Currently, Balancing is in the development and testing phase and is not suitable for production environments.

## Introduction to Balancing

Balancing is a layer 4 load balancing component implemented with eBPF on the Kubernetes platform. It references the functionalities of projects like [cilium](https://github.com/cilium/cilium) , [calico](https://github.com/projectcalico/calico) , and [KPNG](https://github.com/kubernetes-retired/kpng) , and implements load balancing resolution capabilities independent of CNI.

### Current Features

1. **Service resolution inside and outside the cluster**: Replaces kube-proxy.
    - Supports service resolution initiated by Pods and Nodes on cluster nodes based on cGroup eBPF.
    - Supports cGroup eBPF resolution for local applications on external hosts.
    - Future versions will support north-south nodePort resolution on node network cards based on TC eBPF.
    - For more information, please refer to [service](./docs/usages/service.md])

2. **localRedirect policy layer 4 load balancing resolution**:
    - Provides service redirection resolution initiated by Pods and Nodes based on cGroup eBPF.
    - Typical scenarios include redirecting application requests to coreDns to the local coreDns.
    - For more information, please refer to [LocalRedirect Policy](./docs/usages/localredirect.md)

3. **balancing policy layer 4 load balancing resolution (in progress)**:
    - Implements custom global layer 4 load balancing resolution for Pods, Nodes, and applications based on cGroup eBPF.
    - Application scenarios include client-side load balancing resolution for external hosts and load balancing within Kubernetes clusters.
    - For more information, please refer to  [Balancing Policy](./docs/usages/balancing.md)
    - Note: The balancing policy currently only implements load balancing resolution and has not yet implemented inter-node tunnel communication.

4. **Event Logging**:
    - Records load balancing resolution events and associates related container information.

## Typical Use Cases

1. **Service resolution replacing kube-proxy**:
    - Suitable for overlay CNI, such as [Macvlan](https://github.com/containernetworking/plugins/tree/main/plugins/main/macvlan) , [SR-IOV CNI](https://github.com/k8snetworkplumbingwg/sriov-cni) , etc.

2. **coreDns redirection to Node-local DNS**:
    - Provides high-availability redirection forwarding.

3. **External applications accessing Kubernetes services**:
    - Access services within the Kubernetes cluster through layer 4 load balancing addresses.

4. **Layer 4 load balancing access between multiple clusters**.

## Architecture

![arch](./docs/images/arch.png)

The Balancing component consists of an agent and a controller:
- **controller deployment**: Performs webhook validation and modification of various CRD objects.
- **agent daemonset**: Loads eBPF programs and distributes forwarding rules.

![eBPF](./docs/images/cgroup-ebpf.png)

## Quick Start

- Refer to [Installation](./docs/usages/install.md) for quick deployment.
- Refer to [Service Resolution](./docs/usages/service.md) for usage experience.
- Refer to [LocalRedirect Policy](./docs/usages/localredirect.md) for usage experience.
- Refer to [Balancing Policy](./docs/usages/balancing.md) for usage experience.

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
