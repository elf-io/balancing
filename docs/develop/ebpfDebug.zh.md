# eBPF 调试

在部署 Balancing 后，可检测如下内容，确认 Balancing 工作符合预期。


## eBPF map 作用

```
map_service
	记录了为每一个 service 生成的访问入口规则，为 port、nodeport、Loadbalancer、externalIP 分别都会生成一条记录 

map_backend
	记录了每一个后端 endpoint 的转发地址，每一个 endpoint 都会有一个记录，如果是nodePort 场景，记录数量再翻倍

map_affinity
	存储客户端亲和记录，实现会话亲和性

map_nat_record
	记录了 nat 链路追踪

map_node_proxy_ip
	记录了 节点的 id 和 ip 的映射 , 用于其它表格在查询时进行 id 和 ip 之间的索引

map_node_ip
	记录了每一个 node 的 ip ，用于匹配 nodePort 场景下的 目的 ip

map_configure
	ebpf 程序的实时工作配置
```

## 节点 eBPF 检查

```bash
# 在主机如下目录，挂载了 eBPF map
ls /sys/fs/bpf/balancing/
# 输出示例
map_affinity  map_backend  map_configure  map_event  map_nat_record  map_node_ip  map_node_proxy_ip  map_service

# 使用如下命令能够查询到 balancing eBPF map
bpftool map

# 主机如下目录挂载了 cgroup v2
ls /sys/fs/cgroup
# 输出示例
cgroup.controllers  cgroup.stat  cpuset.cpus.isolated  dev-mqueue.mount  io.prio.class  memory.reclaim  proc-sys-fs-binfmt_misc.mount  system.slice

# 使用如下命令能够查询到 balancing 加载的 cgroup_sock_addr 类型的 eBPF Program
bpftool prog

# 使用如下命令能够查询到 balancing 把 eBPF Program 关联到了 cgroup v2
bpftool cgroup tree /run/balancing/cgroupv2

# 主机上查看 eBPF 程序打印出的日志
bpftool prog tracelog
# 或
cat /sys/kernel/debug/tracing/trace_pipe
```

## 确认 Balancing agent 日志

```bash
# 查询 agent pod 的负载均衡解析事件日志
kubectl logs -n elf balancing-agent-q727g | grep "formatted ebpf event" | jq .
```

## 对象

如下这些对象的 id， 都是用在 ebpf 的 map 中代表相关对象

```bash
# 确认每一个节点，都被标记了如下 ID 唯一的 annotation
kubectl get nodes -o jsonpath='{.items[*].metadata.annotations}' | jq .
# 输出示例
{
  "balancing.elf.io/nodeId": "596592060",
  "balancing.elf.io/nodeProxyIpv4": "192.168.0.10",
  ...
}
```

```bash
# 所有的 balancingpolicies ，都有一个唯一的 id
kubectl get balancingpolicies -o jsonpath='{.items[*].metadata.annotations}' | jq .
# 输出示例
{
  "balancing.elf.io/serviceId": "20003",
  ...
}
```

```bash
# 所有的 localredirectpolicies ，都有一个唯一的 id
kubectl get localredirectpolicies -o jsonpath='{.items[*].metadata.annotations}' | jq .
# 输出示例
{
  "balancing.elf.io/serviceId": "10091",
  ...
}
```

## 查看 eBPF 中的数据

进入 agent pod 中，可使用 inspect 命令查看 eBPF map 中的数据。

```bash
# 查询所有 ebpf map 中的数据
inspect showMapData all

# 追踪指定 service 相关的 ebpf map 数据
inspect traceMapData service $namespace $serviceName

# 追踪指定 localredirectpolicies 相关的 ebpf map 数据
inspect traceMapData localRedirect $namespace $policyName

# 追踪指定 balancingpolicies 相关的 ebpf map 数据
inspect traceMapData balancing $namespace $serviceName
```
