# eBPF Debugging

After deploying Balancing, you can check the following to ensure that Balancing is working as expected.

## Node eBPF Check

```bash
# The eBPF map is mounted in the following directory on the host
ls /sys/fs/bpf/balancing/
# Example output
map_affinity  map_backend  map_configure  map_event  map_nat_record  map_node_ip  map_node_proxy_ip  map_service

# Use the following command to query the balancing eBPF map
bpftool map

# The cgroup v2 is mounted in the following directory on the host
ls /sys/fs/cgroup
# Example output
cgroup.controllers  cgroup.stat  cpuset.cpus.isolated  dev-mqueue.mount  io.prio.class  memory.reclaim  proc-sys-fs-binfmt_misc.mount  system.slice

# Use the following command to query the eBPF Program loaded by balancing
bpftool prog

# Use the following command to query the association of the eBPF Program with cgroup v2 by balancing
bpftool cgroup tree /run/balancing/cgroupv2

# View the logs printed by the eBPF program on the host
bpftool prog tracelog
# or
cat /sys/kernel/debug/tracing/trace_pipe
```

## Confirm Balancing Agent Logs

```bash
# Query the load balancing parsing event logs of the agent pod
kubectl logs -n elf balancing-agent-q727g | grep "formatted ebpf event" | jq .
```

## Objects

```bash
# Confirm that each node is marked with a unique annotation ID
kubectl get nodes -o jsonpath='{.items[*].metadata.annotations}' | jq .
# Example output
{
  "balancing.elf.io/nodeId": "596592060",
  "balancing.elf.io/nodeProxyIpv4": "192.168.0.10",
  ...
}
```

```bash
# All balancingpolicies have a unique id
kubectl get balancingpolicies -o jsonpath='{.items[*].metadata.annotations}' | jq .
# Example output
{
  "balancing.elf.io/serviceId": "20003",
  ...
}
```

```bash
# All localredirectpolicies have a unique id
kubectl get localredirectpolicies -o jsonpath='{.items[*].metadata.annotations}' | jq .
# Example output
{
  "balancing.elf.io/serviceId": "10091",
  ...
}
```

## View Data in eBPF

Enter the agent pod, and use the inspect command to view data in the eBPF map.

```bash
# Query all data in the ebpf map
inspect showMapData all

# Trace ebpf map data related to a specific service
inspect traceMapData service $namespace $serviceName

# Trace ebpf map data related to a specific localredirectpolicies
inspect traceMapData localRedirect $namespace $policyName

# Trace ebpf map data related to a specific balancingpolicies
inspect traceMapData balancing $namespace $serviceName
```
