# Service 解析

## 介绍

Service 解析功能参考了 [Cilium](https://github.com/cilium/cilium) 、[Calico](https://github.com/projectcalico/calico) 、[KPNG](https://github.com/kubernetes-retired/kpng)  等项目的相关功能实现，提供了 CNI 无关的负载均衡解析扩展能力。

它能够为以下对象发送的请求实施重定向访问：
* Pod 中的应用
* 集群节点上的应用
* 集群外部主机上的应用

## 功能

当前，它具备如下功能：

* [x] 为 Pod 和 Node 完成东西向的 Service 解析：支持它们主动访问 ClusterIP、NodePort、LoadBalancer、ExternalIP，支持基于 ClientIP 的 sessionAffinity，支持 internalTrafficPolicy 值为 Local。
* [x] 为集群外部的主机应用，通用支持 Service 的地址解析，包括 ClusterIP、NodePort、LoadBalancer、ExternalIP。但是，Balancing 解析为 Pod IP 地址，因此，它适用于 Kubernetes 集群中使用了诸如 [Macvlan](https://github.com/containernetworking/plugins/tree/main/plugins/main/macvlan) 、[Spiderpool](https://github.com/spidernet-io/spiderpool)  等 underlay CNI 的场景，也适用于使用了 BGP 传播了集群 Pod 路由的 [Calico](https://github.com/projectcalico/calico) 等场景。

在后续版本中，解决如下问题：

* [ ] 在 Node 上完成南北向的 Service 解析：支持解析集群外部发送的 Service 访问请求，包括 NodePort、LoadBalancer、ExternalIP，支持基于 ClientIP 的 sessionAffinity，支持 externalTrafficPolicy 值为 Local。
* [ ] 对于存量的 sessionAffinity=ClientIP 转发记录，应该遵循 backend Pod 的健康状态，当 backend Pod 不可用时，要中断持久化转发。

## 使用场景

* 相比 iptables 等传统技术，基于 eBPF 技术实现更加优异的 Service 解析性能，避免痛苦的 iptables 排障。

* 它与 CNI 无关，为许多不具备 eBPF 技术的 CNI 项目完成 eBPF Service 解析，例如 [Antrea](https://github.com/antrea-io/antrea) 、 [Kube-ovn](https://github.com/kubeovn/kube-ovn) 、 [Flannel](https://github.com/flannel-io/flannel) 等项目，以及一些公有云 Kubernetes 集群的 [Amazon VPC CNI](https://github.com/aws/amazon-vpc-cni-k8s) 、[Azure CNI](https://github.com/Azure/azure-container-networking) 。并且，它适用于诸多 underlay CNI，解决因为数据包转发路径的天生不能访问 Service 的问题，例如 [Macvlan](https://github.com/containernetworking/plugins/tree/main/plugins/main/macvlan) 、 [SR-IOV CNI](https://github.com/k8snetworkplumbingwg/sriov-cni) 、 [Spiderpool](https://github.com/spidernet-io/spiderpool) 、[OVS-CNI](https://github.com/k8snetworkplumbingwg/ovs-cni) 。

  > [Cilium](https://github.com/cilium/cilium) 、 [Calico](https://github.com/projectcalico/calico) 自带了 kube-proxy 替代功能，不需要使用 Balancing 项目。
