# Balancing Policy

## 简介

Balancing Policy 定义了一种全新的、更加自由的负载均衡模式，形成了对 kubernetes 的 service 一种补充，它实现集群的全局四层负载均衡转发。

它能够为以下对象发送的请求实施重定向访问：
* Pod 中的应用
* 集群节点上的应用
* 集群外部主机上的应用

## 功能

当前，它具备如下功能：
* [x] 自定义负载均衡的 front 地址，它可以是 kubernetes 的 service name，它也可以是自定义的 VIP 和端口
* [x] 自定义负载均衡的 backend 地址，它可以通过 pod label selector 来指定后端的转发对象，而转发的地址支持三种方式：
    * endpoint IP：负载均衡地址被 DNAT 解析为 pod ip，它适用于所有的 POD
    * HostPort: 负载均衡地址被 DNAT 解析为 pod 所在的 node IP 和 pod HostPort，它适用于定义了 hostPort 端口的 POD
    * nodeProxy: 负载均衡地址被 DNAT 解析为节点的 Proxy Ip 和策略中定义的端口，它适用于所有的 POD。其中，节点的 Proxy Ip 是定义在 node 对象的 annotation `"balancing.elf.io/nodeProxyIpv4": "192.168.0.10"`， 可以通过如下两种方式生成：
        * Balancing agent 可以自动在节点上建立隧道接口，把隧道接口更新到 node 的 annotation 中，这种场景使用于多集群之间联通、集群外部主机应用联通。
        * 管理员可以在 Node 对象上书写覆盖 annotation `"balancing.elf.io/nodeProxyIpv4"` 例如，它可以是该节点的代理映射 IP 、公网映射 IP 等

> Balancing Policy 实例之间、 LocalRedirect Policy 实例之间，它们的 front 不支持绑定相同的 service，或者定义相同的虚拟地址，否则会出现解析冲突的问题

> 当 Balancing Policy 或  LocalRedirect Policy 的 front 使用了自定义 IP 地址，如果该 IP 地址和某个 kubernetes 的 service ClusterIP 冲突，那么优先按照 Balancing Policy 或  LocalRedirect Policy 的转发规则来解析

在后续版本中，解决如下问题：
* [ ] Balancing Agent 支持自动在节点间建立转发隧道，把 IP 地址更新到 node 的 annotation `"balancing.elf.io/nodeProxyIpv4"`， 使得在 overlay CNI 场景下实现集群外部的主机应用、多集群之间的通信互联

## 使用场景

1. 集群外部主机、kubevirt 虚拟机、kubedge 边缘节点上运行 balancing agent 二进制或者 docker 服务，访问到 kubernetes 集群中的服务。

   传统的 nodePort 或者 Loadbalancer 负载均衡解析，会遇到 SNAT 的源端口冲突、长连接超时时间不一致等问题，成为高并发访问的瓶颈。balancing 提供的新方案，能够实现客户端侧的负载均衡解析，减少了转发路径，降低排障难度。

    > 当前版本，Balancing 还未完成节点间的隧道建立和端口分配，因此，只能在 underlay CNI 场景下能保障集群内外的联通。在后续版本中，Balancing 完成隧道建立后，才能保障 overlay CNI 场景下的集群内外连通性。

2. 实施多集群之间的四层负载均衡访问

   > 当前版本，Balancing 还未完成节点间的隧道建立和端口分配，因此，只能在 underlay CNI 场景下能保障集群内外的联通。在后续版本中，Balancing 完成隧道建立后，才能保障 overlay CNI 场景下的集群内外连通性。

3. 自定义 front 负载均衡地址、或自定义 backend 的转发地址，来支持更加灵活的负载均衡需求。

## 策略例子

以下例子中，front 指定了 kubernetes 中的某个 service，backend 基于 pod 的 label selector 来转发到 Pod IP

```shell
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-service-podendpoint
spec:
  enabled: true
  frontend:
    serviceMatcher:
      serviceName: http-server-balancing-pod-v4
      namespace: default
      toPorts:
        # the port and protocol must be in line with the service , but the name should not be
        - port: "8080"
          protocol: TCP
          name: p1
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    serviceEndpoint:
      endpointSelector:
        matchLabels:
          app: http-redirect
      # for podEndpoint: the destination IP is podIP, the destination port is pod port
      redirectMode: podEndpoint
      toPorts:
          # port is podPort for podEndpoint mode
        - port: "80"
          protocol: TCP
          name: p1
        - port: "80"
          protocol: TCP
          name: p2
```

以下例子中，front 指定了 kubernetes 中的某个 service，backend 基于 pod 的 label selector 来转发到 Pod 所在节点的 hostPort

```shell
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-service-hostport
spec:
  enabled: true
  frontend:
    serviceMatcher:
      serviceName: http-server-balancing-hostport-v4
      namespace: default
      toPorts:
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    serviceEndpoint:
      endpointSelector:
        matchLabels:
          app: http-redirect
      # for hostPort: the destination IP is node IP, the destination port is hostPort
      redirectMode: hostPort
      toPorts:
          # port is hostPort for hostPort mode
        - port: "20080"
          protocol: TCP
          name: p2
EOF
```

以下例子中，front 指定了 kubernetes 中的某个 service，backend 基于 pod 的 label selector 来转发到 Pod 所在节点的 Proxy IP

```shell
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-service-nodeproxy
spec:
  enabled: true
  frontend:
    serviceMatcher:
      serviceName: http-server-balancing-nodeproxy-v4
      namespace: default
      toPorts:
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    serviceEndpoint:
      endpointSelector:
        matchLabels:
          app: http-redirect
      redirectMode: nodeProxy
      toPorts:
        - port: "20080"
          protocol: TCP
          name: p2
EOF
```

以下例子中，front 使用了自定义的虚拟 IP 和端口，backend 使用了自定义的 IP 和端口

```shell
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-addr
spec:
  enabled: true
  frontend:
    addressMatcher:
      ip: "169.254.169.254"
      toPorts:
        - port: "8080"
          protocol: TCP
          name: p1
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    addressEndpoint:
      addresses:
        - "1.1.1.1"
        - "1.1.1.2"
      toPorts:
        - port: "80"
          protocol: TCP
          name: p1
        - port: "80"
          protocol: TCP
          name: p2
```
