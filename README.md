# elf

1 在东西向实现 kube proxy replacement ， 尤其适用于 spiderpool  ovs-based cni 等 

2 集群外部主机、kubevirt 虚拟机、外部集群 实现 client 端负载均衡，访问集群的 service 
    转发方式支持 pod ip、hostPort、主机隧道（戴实现）

3 支持劫持 service 访问到本地 endpoint，适用于 clusterPedia、local coredns

未来



## 架构 

### controller

1. 为 balancing 和 localredirect 分配一个唯一的 u32 id 
2. 校验 balancing 和 localredirect ，不允许它们之间引用相同的 service 或者 virtual address 

### agent

1. 为每个 node 分配一个唯一的 u32 id 
