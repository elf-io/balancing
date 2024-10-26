# debug

## debug ebpf

```
#主机上的 ebpf map
~# ls /sys/fs/bpf/balancing/
    map_affinity  map_backend  map_configure  map_event  map_nat_record  map_node_ip  map_node_proxy_ip  map_service

#主机上 cgroup v2 挂载
~# ls /sys/fs/cgroup
    cgroup.controllers      cgroup.stat             cpuset.cpus.isolated   dev-mqueue.mount  io.prio.class     memory.reclaim          proc-sys-fs-binfmt_misc.mount  system.slice
    cgroup.max.depth        cgroup.subtree_control  cpuset.mems.effective  init.scope        io.stat           memory.stat             sys-fs-fuse-connections.mount  user.slice
    cgroup.max.descendants  cgroup.threads          cpu.stat               io.cost.model     kubepods.slice    memory.zswap.writeback  sys-kernel-config.mount
    cgroup.pressure         cpu.pressure            cpu.stat.local         io.cost.qos       memory.numa_stat  misc.capacity           sys-kernel-debug.mount
    cgroup.procs            cpuset.cpus.effective   dev-hugepages.mount    io.pressure       memory.pressure   misc.current            sys-kernel-tracing.mount

#查询到相关的
~# bpftool map

~# bpftool prog

~# bpftool cgroup tree /run/balancing/cgroupv2

#主机日志
~#  bpftool prog tracelog
或
~# cat /sys/kernel/debug/tracing/trace_pipe


agent ebpf 访问解析日志
~# kubectl logs -n elf balancing-agent-q727g | grep "formatted ebpf event" | jq .

```


## 对象

```
# 所有节点都有唯一的 annotation id
~# kubectl get nodes -o jsonpath='{.items[*].metadata.annotations}' | jq .
{
  "balancing.elf.io/nodeId": "596592060",
  "balancing.elf.io/nodeProxyIpv4": "192.168.0.10",
  ....
}

```


```
# 所有的 策略，都有一个 唯一的 id
~# kubectl get balancingpolicies -o jsonpath='{.items[*].metadata.annotations}' | jq .
    {
      "balancing.elf.io/serviceId": "20003",
      ...
    }

~# kubectl get localredirectpolicies -o jsonpath='{.items[*].metadata.annotations}' | jq .
    {
      "balancing.elf.io/serviceId": "10091",
      ...
    }

```


##

```
# 查询指定 service 的 ebpf 数据
~# inspect  traceMapData service default redirectserver
		trace the service data of ebpf map for the service default/redirectserver
		
		------------------------------
		map Hash(map_service)#12 :
		    filterNatType service
		    filterSvcV4Id 3385136556
		
		Service Entries:
		[0]: key={ DestIp:172.21.197.201, DestPort:80, protocol:tcp, NatType:service, Scope:0 },
		     value={ SvcId:3385136556, TotalBackendCount:2, LocalBackendCount:1, AffinitySecond:0, NatMode:ServiceClusterIP, ServiceFlags:0, BalancingFlags:0, RedirectFlags:0 }
		account:  1
		
		LocalRedirect Entries:
		
		Balancing Entries:
		
		end map Hash(map_service)#12: account 1
		------------------------------
		
		
		------------------------------
		map Hash(map_backend)#7 :
		    filterNatType service
		    filterSvcV4Id 3385136556
		
		Service Entries:
		[0]: key={ Order:0, SvcId:3385136556, port:80, protocol:tcp, NatType:service, Scope: 0 }
		     value={ PodIp:172.20.235.198 , PodPort:80, NodeId:596592060, NodePort:0 }
		[1]: key={ Order:1, SvcId:3385136556, port:80, protocol:tcp, NatType:service, Scope: 0 }
		     value={ PodIp:172.20.254.131 , PodPort:80, NodeId:79869938, NodePort:0 }
		account:  2
		
		
		
		end map Hash(map_backend)#7: account 2
		------------------------------

```

