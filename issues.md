
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
	- 支持 访问 clusterIP + svcPort

	- loadbalancerIp + svcPort  ( 不支持 loadbalancerIp + nodePort  )

	- externalIP + svcPort ( 不支持 externalIP + nodePort  )

	- nodeIP + nodePort

	- 支持  EXTERNAL_LOCAL_SVC 和 INTERNAL_LOCAL_SVC ，无论是以 cluster ip / nodeport / externalIP，都可实现转发到本地 pod （如果本地没有 后端pod，则解析失败）

	- 支持  affinity ，无论是以 cluster ip / nodeport / externalIP，都可根据 service 中定义的持久化时间进行亲和访问 （ 如果持久化后端 pod 销毁，目前还是会 继续转发 ，需要 增强） 

(2) 支持 crd  localRedirect ， 把访问解析到 client pod 所在的 node 上的 selected pod

	- frontend 支持定义 两种方式 

	     支持解析 service clusterIP （不支持 LoadbalancerIP、externalIP、nodeIP） + crd localRedirect  中的 port  ， 解析为 localPod + crd 中的 toPort

         支持解析 自定义的 virtual IP + virtual port , 解析为 localPod + crd 中的 toPort

	- backend 支持 pod Selector 指向 pod ip


    注：当本地所有 endpoint 挂了， 会继续以 正常的 service 去解析
    // TODO ， qos:   本地 所有 pod 的 connect qos 流控


（3）支持 crd  balancing， 把访问 解析到整个集群范围的 pod 或者 自定义ip 
	
	- frontend 支持定义 两种方式 
				指向 K8S service (只基于 clusterIP )+ service port ， 

				或者 虚拟IP + 虚拟端口

	- backend 支持2种指定方式
			serviceEndpoint 支持 
				redirectMode: podEndpoint ， 指向 pod ip + crd 中的 port
				redirectMode: hostPort ， 指向 node ip + crd 中的 port
				redirectMode: nodeProxy ， 指向 nodeEntry ip （可以是自建的 vxlan 隧道ip，也可以用户自定义 ） + crd 中的 port

			addressEndpoint：自定义 后端 ip 和 端口




注意：
		（1）优先级：
		当请求的目的地址 相同时，按照如下 优先级 生效
		localRedirect > balancing > service
		
		（2）
		localRedirect 和 balancing policy 中不支持 指向相同的 service 或者 virtual ip




============ 问题

高优先级：
	controller 进行限制，只能有一个 BalancingPolicy / LocalRedirectPolicy 绑定 同名 service   , 否则 agent 侧会 相互覆盖数据

    貌似每次启动，calico node 都歇菜了 ？ 测试和 其它 ebpf 程序的 工程兼容性

	支持  affinity ，无论是以 cluster ip / nodeport / externalIP，都可根据 service 中定义的持久化时间进行亲和访问 （ 如果持久化后端 pod 销毁，目前还是会 继续转发 ，需要 增强）

	实现 nodeEntry 中 ，在 主机的 隧道 IP 上 生效 指定 port 的 DNAT

	实现 nodeEntry 中 ，在 主机的 隧道 IP 上 生效 指定 port 的 DNAT


目前只支持 ipv4， 不支持 ipv6

如果 node ip 变换了，目前 backend 中的 pod 所在 的 node ip 不会变化，需要增强

对于识别为 local 的 pod，例如 default/kubernetes 的 endpointslice， 其 yaml 中就不带 nodeName， 导致 识别 失败

程序启动时，会清除 service backend node map， 实现数据完整同步 。 这样，可能会带来 短暂的 service 访问失败

支持 解析ip 的 指标



