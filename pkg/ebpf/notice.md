
=============  应用场景

(1)  k8s 集群内实现 service 解析，包括 clusterip 、 nodePort 等

（2）支持集群外 主机部署， 实现 主机应用 直接访问到 pod ip（macvlan）或者  pod 所在主机 ip + nodePort
这样，能够避免传统 nodePort 等方案带来的 源端口冲突、并发低、转发性能差 等问题

    尤其是 kubevirt 虚拟机场景

(3) 支持 localRedirect，支持 local dns

（4） 为多集群 如 K3S 而服务

（5）支持 kubeedge， 在 边端 不需要 cni 的情况下，边端进行 clusterIP 解析，把流量 封发到 云端
pod 所在节点的 nodePort

============================ 功能


(1) 支持 service 的访问
支持 访问 clusterIP + svcPort
loadbalancerIp + svcPort  ( 不支持 loadbalancerIp + nodePort  )
externalIP + svcPort ( 不支持 externalIP + nodePort  )
nodeIP + nodePort


(2) 支持 crd  localRedirect
onlyLocal:  当本地 endpoint 挂了，是否 允许 正常 访问 service
qos:   本地 所有 pod 的 connect qos 流控

（3）支持 crd  balancing
自定义 浮动 ip  和 后端 endpoint ip
可以 额外 自定义  endpoint ip
也可以关联 K8S 的 service  
fowardToNode ： 是否解析到 pod 所在的 node 的 nodePort  ， 适用与集群外部 的节点


============ 问题

目前只支持 ipv4， 不支持 ipv6

如果 node ip 变换了，目前 backend 中的 pod 所在 的 node ip 不会变化，需要增强

对于识别为 local 的 pod，例如 default/kubernetes 的 endpointslice， 其 yaml 中就不带 nodeName， 导致 识别 失败

程序启动时，会清除 service backend node map， 实现数据完整同步 。 这样，可能会带来 短暂的 service 访问失败 

貌似每次启动，calico node 都歇菜了 ？ 测试和 其它 ebpf 程序的 工程兼容性 

还没测试 udp 


鸡和蛋的问题： 没有 kube-proxy， 我们的组价 部署上来时，如何访问 api-server 进行 工作 ？

支持 crd redirect 和 balancing 

支持 解析ip 的 指标


