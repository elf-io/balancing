

```
#主机上的 ebpf map
~# ls /sys/fs/bpf/balancing/
    map_affinity  map_backend  map_configure  map_event  map_nat_record  map_node_ip  map_node_proxy_ip  map_service

#主机上 cgroup v2 挂载
~# ls /run/balancing

#查询到相关的 ebpf map 和 program
~# bpftool map
~# bpftool cgroup tree /run/balancing

```

