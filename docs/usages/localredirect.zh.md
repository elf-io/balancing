# LocalRedirect Policy

## 介绍

LocalRedirect Policy，它参考了 [cilium](https://github.com/cilium/cilium) 项目的相关功能实现，基于 cGroup eBPF 技术，在 Pod 访问指定 service 时，重定向解析到同节点的本地服务。

它能够为以下对象发送的请求实施重定向访问：
* Pod 中的应用
* 集群节点上的应用

![redirect](../images/localRedirect.png)

## 功能

当前版本，支持如下功能：

* [x] front 支持指向 service，也支持指向自定义的 VIP 和端口
* [x] backend 支持 pod label selector

> Balancing Policy 实例之间、 LocalRedirect Policy 实例之间，它们的 front 不支持绑定相同的 service，或者定义相同的虚拟地址，否则会出现解析冲突的问题

> 当 Balancing Policy 或  LocalRedirect Policy 的 front 使用了自定义 IP 地址，如果该 IP 地址和某个 kubernetes 的 service ClusterIP 冲突，那么优先按照 Balancing Policy 或  LocalRedirect Policy 的转发规则来解析

## 使用场景

* 为 Node-local DNS 实施透明的重定向
  为了提高 coreDns 的服务能力，避免 DNS 雪崩效应，开源社区引入了 Node-local DNS ，完成 per-Node 的 DNS 缓存。

  传统方式中，修改 pod 的 DNS 配置，指向本地的 Node-local DNS 的虚拟地址，在节点上绑定了该虚拟地址，因此，在 Node-local DNS 发生故障或者升级时，这并不能为应用完成高可用的 DNS 服务。

  引入 LocalRedirect Policy 重定向能力，在不需要引入 pod 的 DNS 配置和新的虚拟地址情况下，为 Node-local DNS 提供了透明的、高可用的服务重定向，支持本地的  Node-local DNS 故障或者升级过程中，服务访问解析到原本的 coreDNs 服务。

## 策略例子

如下 yaml 例子，front 指向 service name

```shell
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
        # 只能有一个 name: p1
        - port: "8080"
          protocol: TCP
          name: p1
        # 只能有一个 name: p2
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    endpointSelector:
      matchLabels:
        app: http-redirect
    toPorts:
        # 只能有一个 name: p1
      - port: "80"
        protocol: TCP
        name: p1
        # 只能有一个 name: p2
      - port: "80"
        protocol: TCP
        name: p2
```

如下 yaml 例子，front 使用自定义的虚拟 IP 和端口

```shell
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
        # 只能有一个 name: p1
        - port: "8080"
          protocol: TCP
          name: p1
        # 只能有一个 name: p2
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
