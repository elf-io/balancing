

## map 结构 

一个 service 会生成多组 service 数据 和 多组 backend 数据 

以一个 Nodeport service 为例，有如下数据 

service 数据如下，它有 两个 port 转发的定义 
```shell
default       http-server-v4            Nodeport   172.26.75.227    <pending>         80:31399/TCP,8080:31633/TCP   2m3s
```

所有数据，会被 相同的 SvcId:3813350060  串联在一起

在 service map 中，有如下数据， 它的数量是 ： "service 端口数" * ( clusterIP数量 + LoadbalancerIP数量 + externalIP数量 + 一个 nodePort  )
```shell
# 对应 80 端口 ： ClusterIP + servicePort
5 : key={ DestIp:172.26.75.227, DestPort:80, protocol:tcp, NatType:service, Scope:0 }
5 : value={ SvcId:3813350060, TotalBackendCount:2, LocalBackendCount:1, AffinitySecond:0, NatMode:ServiceClusterIP, ServiceFlags:0, BalancingFlags:0, RedirectFlags:0 }

# 对应 80 端口 ：  nodeIP + nodePort :（ 这种数据包的转发，去查询  node map 来确认 node ip ）
7 : key={ DestIp:255.255.255.255, DestPort:31399, protocol:tcp, NatType:service, Scope:0 }
7 : value={ SvcId:3813350060, TotalBackendCount:2, LocalBackendCount:1, AffinitySecond:0, NatMode:ServiceNodePort, ServiceFlags:0, BalancingFlags:0, RedirectFlags:0 }

# 对应 8080 端口 ： ClusterIP + servicePort
11 : key={ DestIp:255.255.255.255, DestPort:31633, protocol:tcp, NatType:service, Scope:0 }
11 : value={ SvcId:3813350060, TotalBackendCount:2, LocalBackendCount:1, AffinitySecond:0, NatMode:ServiceClusterIP, ServiceFlags:0, BalancingFlags:0, RedirectFlags:0 }

# 对应 8080 端口 ： loadbalancerIP + servicePort
14 : key={ DestIp:172.26.75.227, DestPort:80, protocol:tcp, NatType:service, Scope:0 }
14 : value={ SvcId:3813350060, TotalBackendCount:2, LocalBackendCount:1, AffinitySecond:0, NatMode:ServiceLoadBalancerIP, ServiceFlags:0, BalancingFlags:0, RedirectFlags:0 }

#.... 可能还有 externalIP 的记录
#.... 可能还有 loadbalancer ip 的记录

```

在 backend map 中，有如下数据， 记录的数量是：   
		如果有没有 nodePort ，记录数量是： endpoint数量 * service 端口数量
		如果有 nodePort ，   记录数量是： endpoint数量 * service 端口数量 * 2

```shell

# 对应 port 80 端口
0 : key={ Order:0, SvcId:3813350060, port:8080, protocol:tcp, NatType:service, Scope: 0 }
0 : value={ PodIp:172.25.161.5 , PodPort:80, NodeId:xxxxxx, NodePort:31633 }
10 : key={ Order:1, SvcId:3813350060, port:8080, protocol:tcp, NatType:service, Scope: 0 }
10 : value={ PodIp:172.25.132.2 , PodPort:80, NodeId:xxxxxx, NodePort:31633 }
# 对应 nodePort 31633 端口
1 : key={ Order:0, SvcId:3813350060, port:32431, protocol:tcp, NatType:service, Scope: 0 }
1 : value={ PodIp:172.25.161.5 , PodPort:80, NodeId:xxxxxx, NodePort:31633 }
11 : key={ Order:1, SvcId:3813350060, port:32431, protocol:tcp, NatType:service, Scope: 0 }
11 : value={ PodIp:172.25.132.2 , PodPort:80, NodeId:xxxxxx, NodePort:31633 }

# 对应 8080 端口
1 : key={ Order:0, SvcId:3813350060, port:80, protocol:tcp, NatType:service, Scope: 0 }
1 : value={ PodIp:172.25.161.5 , PodPort:80, NodeId:xxxxxx, NodePort:31399 }
11 : key={ Order:1, SvcId:3813350060, port:80, protocol:tcp, NatType:service, Scope: 0 }
11 : value={ PodIp:172.25.132.2 , PodPort:80, NodeId:xxxxxx, NodePort:31399 }
# 对应 nodePort 31399 端口
2 : key={ Order:0, SvcId:3813350060, port:31399, protocol:tcp, NatType:service, Scope: 0 }
2 : value={ PodIp:172.25.161.5 , PodPort:80, NodeId:xxxxxx, NodePort:31399 }
22 : key={ Order:1, SvcId:3813350060, port:31399, protocol:tcp, NatType:service, Scope: 0 }
22 : value={ PodIp:172.25.132.2 , PodPort:80, NodeId:xxxxxx, NodePort:31399 }
```



