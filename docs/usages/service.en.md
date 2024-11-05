# Service Resolution

## Introduction

The Service resolution feature is inspired by projects like [Cilium](https://github.com/cilium/cilium) , [Calico](https://github.com/projectcalico/calico) , and [KPNG](https://github.com/kubernetes-retired/kpng) . It provides CNI-independent load balancing resolution capabilities.

It can redirect requests sent by the following objects:
* Applications within Pods
* Applications on cluster nodes
* Applications on external hosts

## Features

Currently, it has the following features:

* [x] East-west Service resolution for Pods and Nodes: Supports accessing ClusterIP, NodePort, LoadBalancer, ExternalIP, supports sessionAffinity based on ClientIP, and supports internalTrafficPolicy set to Local.
* [x] For applications on external hosts, it generally supports Service address resolution, including ClusterIP, NodePort, LoadBalancer, ExternalIP. However, Balancing resolves to Pod IP addresses, making it suitable for scenarios using underlay CNIs like [Macvlan](https://github.com/containernetworking/plugins/tree/main/plugins/main/macvlan) , [Spiderpool](https://github.com/spidernet-io/spiderpool) , or scenarios using [Calico](https://github.com/projectcalico/calico) with BGP propagated cluster Pod routes.

In future versions, the following issues will be addressed:

* [ ] North-south Service resolution on Nodes: Supports resolving Service access requests sent from outside the cluster, including NodePort, LoadBalancer, ExternalIP, supports sessionAffinity based on ClientIP, and supports externalTrafficPolicy set to Local.
* [ ] For existing sessionAffinity=ClientIP forwarding records, it should follow the health status of backend Pods, and persistent forwarding should be interrupted when a backend Pod is unavailable.

## Use Cases

* Compared to traditional technologies like iptables, the eBPF-based implementation offers superior Service resolution performance, avoiding the troubleshooting pain of iptables.

* It is CNI-independent, enabling eBPF Service resolution for many CNI projects that do not have eBPF technology, such as [Antrea](https://github.com/antrea-io/antrea) , [Kube-ovn](https://github.com/kubeovn/kube-ovn) , [Flannel](https://github.com/flannel-io/flannel) , and some public cloud Kubernetes clusters like [Amazon VPC CNI](https://github.com/aws/amazon-vpc-cni-k8s) , [Azure CNI](https://github.com/Azure/azure-container-networking) . Additionally, it is suitable for many underlay CNIs, solving the inherent issue of not being able to access Services due to the data packet forwarding path, such as [Macvlan](https://github.com/containernetworking/plugins/tree/main/plugins/main/macvlan) , [SR-IOV CNI](https://github.com/k8snetworkplumbingwg/sriov-cni) , [Spiderpool](https://github.com/spidernet-io/spiderpool) , [OVS-CNI](https://github.com/k8snetworkplumbingwg/ovs-cni) .

  > [Cilium](https://github.com/cilium/cilium) and [Calico](https://github.com/projectcalico/calico) come with kube-proxy replacement functionality and do not require the Balancing project.
