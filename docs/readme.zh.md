# Balancing

目前，Balancing 处于开发测试阶段，不适合用于生产环境。

## Balancing 简介

Balancing 是一个在 Kubernetes 平台上基于 eBPF 实现的四层负载均衡组件。它参考了 [cilium](https://github.com/cilium/cilium) 、[calico](https://github.com/projectcalico/calico) 、[KPNG](https://github.com/kubernetes-retired/kpng)  等项目的功能，实现了与 CNI 无关的负载均衡解析扩展能力。

### 当前功能

1. **集群内外的 service 解析**：替代 kube-proxy。
    - 支持在集群节点上基于 cGroup eBPF 为 Pod 和 Node 发起的 service 解析。
    - 支持在集群外部主机上为本地应用提供 cGroup eBPF 解析。
    - 未来版本将支持在节点网卡上基于 TC eBPF 实现南北向的 nodePort 解析。
    - 更多信息，请参考 [service解析](./usages/service.zh.md)

2. **localRedirect policy 四层负载均衡解析**：
    - 基于 cGroup eBPF 为 Pod 和 Node 发起的 service 重定向解析。
    - 典型场景如将应用访问 coreDns 的请求重定向到本地的 local coreDns。
    - 更多信息，请参考 [LocalRedirect Policy](./usages/localredirect.zh.md)

3. **balancing policy 四层负载均衡解析（进行中）**：
    - 基于 cGroup eBPF 为 Pod、Node、application 实施自定义的全局四层负载均衡解析。
    - 应用场景包括集群外部主机的客户端侧负载均衡解析、Kubernetes 集群内的负载均衡等。
    - 更多信息，请参考 [Balancing Policy](./usages/balancing.zh.md)
    - 注意：balancing policy 目前仅实现了负载均衡解析，尚未实现节点间的隧道通信。

4. **事件日志**：
    - 记录负载均衡解析事件，并关联相关容器信息。

## 典型使用场景

1. **替代 kube-proxy 的 service 解析**：
    - 适用于 overlay CNI，如 [Macvlan](https://github.com/containernetworking/plugins/tree/main/plugins/main/macvlan) 、[SR-IOV CNI](https://github.com/k8snetworkplumbingwg/sriov-cni)  等。

2. **coreDns 重定向到 Node-local DNS**：
    - 提供高可用的重定向转发。

3. **集群外部应用访问 Kubernetes 服务**：
    - 通过四层负载均衡地址访问 Kubernetes 集群中的服务。

4. **多集群之间的四层负载均衡访问**。

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
