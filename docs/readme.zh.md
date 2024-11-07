# Balancing

目前，Balancing 处于开发测试阶段，不适合用于生产环境。

## Balancing 简介

Balancing 是一个在 Kubernetes 平台上基于 eBPF 实现的四层负载均衡组件。它参考了 [cilium](https://github.com/cilium/cilium) 、[calico](https://github.com/projectcalico/calico) 、[KPNG](https://github.com/kubernetes-retired/kpng)  等项目的功能，
Balancing 支持以容器化方式运行在 Kubernetes 集群内部，也支持二进制方式运行在裸金属上，为 Kubernetes 集群内部应用、外部应用实现了与 CNI 无关的负载均衡访问能力。

### 当前功能

1. **实施 CNI 无关的集群 service 解析**
    - 基于 cGroup eBPF ，在集群节点上为 Pod 和 Node 发起的 service 访问实施解析，实现 kube-proxy replacement。
    - 在集群外的裸金属、虚拟机上实施客户端负载均衡解析，以支持访问 kubernetes 集群中的 service。
    - 未来版本将支持在节点网卡上基于 TC eBPF 实现南北向的 nodePort 解析。
    - 更多信息，请参考 [service解析](./usages/service.zh.md)

2. **实施本地重定向的四层负载均衡解析**：
    - 基于 cGroup eBPF，为 Pod 和 Node 发起的 service 访问重定向解析到同节点上的服务，典型场景如 local coreDns。
    - 更多信息，请参考 [LocalRedirect Policy](./usages/localredirect.zh.md)

3. **实施集群全局、集群外部服务的四层负载均衡解析**：
    - 支持更加灵活的策略定义，为集群内部和外部的应用提供了全局的负载均衡策略。
    - 应用场景包括集群外部主机的客户端侧负载均衡解析、Kubernetes 集群内的负载均衡等。
    - 更多信息，请参考 [Balancing Policy](./usages/balancing.zh.md)

4. **解析事件的指标日志**：
    - 记录负载均衡解析事件，并关联容器信息，形成完整的负载均衡解析指标。

## 典型使用场景

1. **在 CNI 无关的集群中，替代 kube-proxy 的 service 解析**：
    - 适用于无法实施 service 的 underlay CNI，如 [Macvlan](https://github.com/containernetworking/plugins/tree/main/plugins/main/macvlan) 、[SR-IOV CNI](https://github.com/k8snetworkplumbingwg/sriov-cni)  等。
    - 适用于没有实现 eBPF 功能的CNI，如 [Antrea](https://github.com/antrea-io/antrea) 、 [Kube-ovn](https://github.com/kubeovn/kube-ovn) 、 [Flannel](https://github.com/flannel-io/flannel) ，以及公有云集群。

2. **服务访问重定向到本地代理**：
    - 实施高可用的重定向， coreDns 服务定向到 Node-local DNS
    - 为 [clusterpedia](https://github.com/clusterpedia-io/clusterpedia) 实施节点 api-server 代理

3. **集群外部的客户端应用侧实施 eBPF 四层负载均衡，访问 Kubernetes 集群中的服务**：
    - 裸金属、虚拟机中的应用
    - [kubevirt](https://github.com/kubevirt/kubevirt) 虚拟机中的应用
    - [kubeedge](https://github.com/kubeedge/kubeedge) 边端节点访问云端服务（进行中）

4. **多集群之间的四层负载均衡访问**。
    - 跨集群的服务访问 （进行中）

5. **通过自定义的前端和后端地址，在 kubernetes 集群中提供外部主机服务的负载均衡访问**
    - 通过自定义的负载均衡策略，为集群外部的应用提供集群内部的负载均衡访问地址，并实施健康检查（进行中）

## 架构

![arch](./images/arch.png)

Balancing 组件由 agent 和 controller 构成：
- **controller deployment**：实施各种 CRD 对象的 webhook 校验和修改。
- **agent daemonset**：加载 eBPF 程序并下发转发规则。

![eBPF](./images/cgroup-ebpf.png)

## 快速开始

- 参考 [安装](./usages/install.zh.md) 快速部署。
- 参考 [service 解析](./usages/service.zh.md) 进行使用体验。
- 参考 [LocalRedirect Policy](./usages/localredirect.zh.md) 进行使用体验。
- 参考 [Balancing Policy](./usages/balancing.zh.md) 进行使用体验。

## 路线图

- **IP 家族和协议**
  - [x] 支持 TCP 和 UDP
  - [x] 支持 IPv4
  - [ ] 支持 IPv6

- **可观测性**
  - [x] 负载均衡解析日志
  - [ ] 负载均衡解析指标

- **service 解析**
  - [x] 东西向 service 解析
  - [ ] 南北向 service 解析
  - [ ] sessionAffinity 转发记录健康状态

- **LocalRedirect Policy**
  - [x] front 支持指向 service 和自定义 VIP
  - [x] backend 支持 pod 标签选择器

- **Balancing Policy**
  - [x] front 支持指向 service 和自定义 VIP
  - [x] backend 支持 pod 标签选择器
  - [x] backend 支持自定义 IP 和端口
  - [ ] 节点间转发隧道

- **多集群互联**
  - [ ] 跨集群的 service 互联
  - [ ] 跨集群的 balancing policy

- **其它**
  - [x] 支持 amd 架构
  - [ ] 支持 arm 架构

## 许可证

Balancing 遵循 Apache License, Version 2.0 许可协议。详见 [LICENSE](./LICENSE)。
