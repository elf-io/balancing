# service 解析

## 介绍

service 解析功能，它参考了 [cilium](https://github.com/cilium/cilium) 、[calico](https://github.com/projectcalico/calico) 、[KPNG](https://github.com/kubernetes-retired/kpng) 等项目的相关功能实现，实现了 CNI 无关的负载均衡解析扩展能力。

它能够为以下对象发送的请求实施重定向访问：
* Pod 中的应用
* 集群节点上的应用
* 集群外部主机上的应用

## 功能

当前，它具备如下功能：

* [x] 为 POD 和 Node 完成东西向的 service 解析：支持它们主动访问 ClusterIP、NodePort、Loadbalancer、ExternalIp，支持基于 ClientIP 的 sessionAffinity，支持 internalTrafficPolicy 值为 Local
* [x] 为集群外部的主机应用，通用支持 service 的地址解析，包括 ClusterIP、NodePort、Loadbalancer、ExternalIp，但是，Balancing 解析为 Pod IP 地址，因此，它适用于 kubernetes 集群中使用了诸如 [Macvlan](https://github.com/containernetworking/plugins/tree/main/plugins/main/macvlan) 、 [Spiderpool](https://github.com/spidernet-io/spiderpool) 等 underlay CNI 的场景，
    也使用于使用了 BGP 传播了集群 POD 路由的 [calico](https://github.com/projectcalico/calico) 等场景 

在后续版本中，解决如下问题：

* [ ] 在 node 上完成南北向的 service 解析：支持解析集群外部发送的 service 访问请求，包括 NodePort、Loadbalancer、ExternalIp，支持基于 ClientIP 的 sessionAffinity，支持 externalTrafficPolicy 值为 Local
* [ ] 对于存量的 sessionAffinity=ClientIP 转发记录，应该遵循 backend Pod 的健康状态，当 backend Pod 不可用时，要中断持久化转发

## 使用场景

* 相比 iptables 等传统技术，基于 eBPF 技术实现更加优异的 service 解析性能，避免痛苦的 iptables 排障。

* 它与CNI 无关，为许多不具备 eBPF 技术的 CNI 项目完成 eBPF service 解析，例如
  [Antrea](https://github.com/antrea-io/antrea) 、[Kube-ovn](https://github.com/kubeovn/kube-ovn) 、[flannel](https://github.com/flannel-io/flannel) 等项目，
  例如一些公有云 kubernetes 集群的 [amazon-vpc-cni](https://github.com/aws/amazon-vpc-cni-k8s) 、 [azure cni](https://github.com/Azure/azure-container-networking) 。
  并且，它适用于诸多 underlay CNI，解决因为数据包转发路径的天生不能访问 service，例如 [Macvlan](https://github.com/containernetworking/plugins/tree/main/plugins/main/macvlan) 、[SR-IOV CNI](https://github.com/k8snetworkplumbingwg/sriov-cni) 、 [Spiderpool](https://github.com/spidernet-io/spiderpool) 、[ovs-cni](https://github.com/k8snetworkplumbingwg/ovs-cni)

  > [cilium](https://github.com/cilium/cilium) 、[calico](https://github.com/projectcalico/calico) 自带了 kube-proxy replacement 功能，不需要使用 Balancing 项目
