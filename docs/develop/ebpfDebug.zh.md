# eBPF 调试

在部署 Balancing 后，可检测如下内容，确认 Balancing 工作符合预期

## 节点 eBPF 检查

```
# 在主机如下目录，挂载了 eBPF map
~# ls /sys/fs/bpf/balancing/
    map_affinity  map_backend  map_configure  map_event  map_nat_record  map_node_ip  map_node_proxy_ip  map_service

# 使用如下命令能够查询到 balancing eBPF map
~# bpftool map

# 主机如下目录挂载了 cgroup v2
~# ls /sys/fs/cgroup
    cgroup.controllers      cgroup.stat             cpuset.cpus.isolated   dev-mqueue.mount  io.prio.class     memory.reclaim          proc-sys-fs-binfmt_misc.mount  system.slice
    cgroup.max.depth        cgroup.subtree_control  cpuset.mems.effective  init.scope        io.stat           memory.stat             sys-fs-fuse-connections.mount  user.slice
    cgroup.max.descendants  cgroup.threads          cpu.stat               io.cost.model     kubepods.slice    memory.zswap.writeback  sys-kernel-config.mount
    cgroup.pressure         cpu.pressure            cpu.stat.local         io.cost.qos       memory.numa_stat  misc.capacity           sys-kernel-debug.mount
    cgroup.procs            cpuset.cpus.effective   dev-hugepages.mount    io.pressure       memory.pressure   misc.current            sys-kernel-tracing.mount

# 使用如下命令能够查询到 balancing 加载的 eBPF Program
~# bpftool prog

# 使用如下命令能够查询到 balancing 把 eBPF Program 关联到了 cgroup v2
~# bpftool cgroup tree /run/balancing/cgroupv2

# 主机上查看 eBPF 程序打印出的日志
~#  bpftool prog tracelog
# 或
~# cat /sys/kernel/debug/tracing/trace_pipe

```

## 确认 Balancng agent 日志

```
# 查询 agent pod 的负载均衡解析事件日志
~# kubectl logs -n elf balancing-agent-q727g | grep "formatted ebpf event" | jq .
```

## 对象

```
# 确认每一个节点，都被标记了如下 ID 唯一的 annotation
~# kubectl get nodes -o jsonpath='{.items[*].metadata.annotations}' | jq .
{
  "balancing.elf.io/nodeId": "596592060",
  "balancing.elf.io/nodeProxyIpv4": "192.168.0.10",
  ....
}

```

```
# 所有的 balancingpolicies ，都有一个唯一的 id
~# kubectl get balancingpolicies -o jsonpath='{.items[*].metadata.annotations}' | jq .
    {
      "balancing.elf.io/serviceId": "20003",
      ...
    }
```

```
# 所有的 localredirectpolicies ，都有一个唯一的 id
~# kubectl get localredirectpolicies -o jsonpath='{.items[*].metadata.annotations}' | jq .
    {
      "balancing.elf.io/serviceId": "10091",
      ...
    }

```

## 查看 eBPF 中的数据

进入 agent pod 中，可使用 inspect 命令查看 eBPF map 中的数据 

```

# 查询所有 ebpf map 中的数据
~# inspect showMapData all

# 追踪指定 service 相关的 ebpf map 数据
~# inspect traceMapData service $namespace $serviceName

# 追踪指定 localredirectpolicies 相关的 ebpf map 数据
~# inspect traceMapData localRedirect $namespace $policyName

# 追踪指定 balancingpolicies 相关的 ebpf map 数据
~# inspect traceMapData balancing $namespace $serviceName

```
