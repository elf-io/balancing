#ifndef __MAP_H__
#define __MAP_H__


#include "vmlinux.h"


//======================================= map ： 存储 service ， 包括  service( clusterip nodePort  loadbalancer ) localRedirect  loadbalancer

#define DEFAULT_MAX_EBPF_MAP_ENTRIES 65536

// map 的 key 和 value struct，会自动在 golang 中 生成 struct

#ifdef ENABLE_IPV4

#define NAT_TYPE_SERVICE    0
#define NAT_TYPE_REDIRECT   1
#define NAT_TYPE_BALANCING    2

#define NODE_PORT_IP	0xffffffff

/*
    对于 NAT_TYPE_SERVICE ：
        无论目的是 clusterIP，nodePort，LoadBalancer  ， 都是解析到 pod ip
    对于 NAT_TYPE_REDIRECT：
        解析到 本地 pod ip，
*/
// a loadbalancer has several entries mapping to each backend
struct mapkey_service {
  __be32 address;     /* 小端存储。 clusterIP， 或者 NODE_PORT_IP(255.255.255.255) 表示 nodePort  */
  __be16 dport;            /* 小端存储。 clusterIP 的 端口， 或者 nodePort 的 端口     */
  __u8  proto;
  __u8  nat_type;         /* NAT_TYPE_SERVICE (  lowest priority  ) ,NAT_TYPE_REDIRECT ,  NAT_TYPE_BALANCING ( highest priority )  */
  __u8  scope;
  __u8  pad[3];
};


#define SERVICE_FLAG_EXTERNAL_LOCAL_SVC	0x1
#define SERVICE_FLAG_INTERNAL_LOCAL_SVC	0x2
//
//
#define NAT_FLAG_ALLOW_ACCESS_SERVICE	0x1


#define NAT_MODE_SERVICE_CLUSTERIP	    0x1
#define NAT_MODE_SERVICE_LOADBALANCER	0x2
#define NAT_MODE_SERVICE_EXTERNALIP  	0x3
#define NAT_MODE_SERVICE_NODEPORT   	0x4
//
#define NAT_MODE_REDIRECT_SERVICE   	        0x10
#define NAT_MODE_REDIRECT_ADDRESS   	        0x11
//
#define NAT_MODE_BALANCING_SERVICE_POD   	        0x20
#define NAT_MODE_BALANCING_SERVICE_HOSTPORT   	    0x21
#define NAT_MODE_BALANCING_SERVICE_NODEPROXY   	    0x22
#define NAT_MODE_BALANCING_ADDRESS   	            0x23


struct mapvalue_service {
  __u32 svc_id ;                 // 一个 service 有一个 唯一的 ID ，用来映射 service 下 所有的 endpoint
  __u32 total_backend_count;         // how many global backend exist in the service
  __u32 local_backend_count;         // how many local-node backend exist in the service ，用于实现 clientIP 亲和
  __u32 affinity_second;       /* In seconds, only for svc frontend */
  __u8  service_flags;                /* SERVICE_FLAG_EXTERNAL_LOCAL_SVC  , SERVICE_FLAG_INTERNAL_LOCAL_SVC */
  __u8  balancing_flags;                /* not used  */
  __u8  redirect_flags;         /* NAT_FLAG_ALLOW_ACCESS_SERVICE( 如果在 local-node backend 不可用时，是否正常解析到 clusterIP)  */
  __u8  nat_mode;               /* 用于标识 对应的 key 中的 ip 地址是 哪种哪些， 参考 NAT_MODE_* 标志 */
};

// 一个 k8s 的 service， 若其下有 n(默认1) 个 clusterIP ，有 m 个 loabdlancerIP， 有 t 个 externalIP， 没有 nodePort ， 那么  map_service 中有 n+m+t 个记录
// 一个 k8s 的 service， 若其下有 n(默认1) 个 clusterIP ，有 m 个 loabdlancerIP， 有 t 个 externalIP， 有 nodePort ， 那么  map_service 中有 n+m+t+1 个记录
struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, struct mapkey_service  );
  __type(value, struct mapvalue_service  );
  __uint(pinning, 1); /* 这个配合 golang 中的 pinPath，完成 路径 pin */
  //__uint(map_flags, 0);
  __uint(max_entries, DEFAULT_MAX_EBPF_MAP_ENTRIES);
} map_service SEC(".maps");


//======================================= map ： 存储 endpoint  ， pod ip

struct mapkey_backend {
    __u32  svc_id;  // 对应 mapvalue_service 中的 svc_id ，  一个 service 有一个 唯一的 ID ，用来映射 service 下 所有的 endpoint
    __be32 order;      //  第几个 endpoint ip 。 前面几个记录，优先存储 本地 node 上的 endpoint ， 用于实现 clientIP 亲和
    __be16 dport;
    __u8   proto;
    __u8   nat_type;
    __u8   scope;
    __u8   pad[3];
};

struct mapvalue_backend {
	__be32 pod_address;		/* 小端存储。 Service endpoint IPv4 address , saved in LittleEndian */
	__u32 node_id;		    /* 对应 map_node_entry 的 key，查询 node 的 入口 ip 。 小端存储。 for loadbalancer , access the nodePort */
	__be16 pod_port;		/* 小端存储。 L4 port filter , saved in LittleEndian */
	__be16 node_port;		/* 小端存储。 for loadbalancer , access the nodePort */
};

// 一个 k8s 的 service， 若对应有 n 个 endpoint ， 该 service 下有 m 个 port，每个 port 没有 nodePort， 那么  map_backend 中有 n*m 个记录
// 一个 k8s 的 service， 若对应有 n 个 endpoint ， 该 service 下有 m 个 port，每个 port 有 nodePort，   那么  map_backend 中有 n*m*2 个记录
struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, struct mapkey_backend );
  __type(value, struct mapvalue_backend );
  __uint(max_entries, DEFAULT_MAX_EBPF_MAP_ENTRIES);
  __uint(pinning, 1); /* 这个配合 golang 中的 pinPath，完成 路径 pin */
  //__uint(map_flags, 0);
} map_backend SEC(".maps");


//-----------------------  map ： key 存储 所有 node ipv4 ip ，用于匹配 nodePort 中的 主机 ip ； value 未使用

typedef __u32 __ipv4_ip ;

struct {
  __uint(type, BPF_MAP_TYPE_HASH );
  __type(key, __ipv4_ip  );  // node ipv4 ip
  __type(value, __u32  );  // 未使用
  __uint(max_entries, DEFAULT_MAX_EBPF_MAP_ENTRIES);
  __uint(pinning, 1); /* 这个配合 golang 中的 pinPath，完成 路径 pin */
  //__uint(map_flags, 0);
} map_node_ip SEC(".maps");


//-----------------------  map ： key 存储 每一个 node 的唯一 id ，value 存储 node 的 入口 ip

struct {
  __uint(type, BPF_MAP_TYPE_HASH );
  __type(key,   __u32  );  // node id
  __type(value, __ipv4_ip  );   // node ipv4 ip ， 可以是 node ip， 或者 node 的 隧道ip，或者 node 的公网映射 ip
  __uint(max_entries, DEFAULT_MAX_EBPF_MAP_ENTRIES);
  __uint(pinning, 1); /* 这个配合 golang 中的 pinPath，完成 路径 pin */
  //__uint(map_flags, 0);
} map_node_proxy_ip SEC(".maps");


//======================================= map ： 存储 亲和 记录

struct mapkey_affinity {
   __u64   client_cookie;       //  bpf_get_netns_cookie(ctx);
   __be32  original_dest_ip;  // 小端存储。
   __u16   original_port ;   // 小端存储。
   __u8    proto;
   __u8    pad;
};

struct mapvalue_affinity {
   __u64    ts;             // 这个值存储了 上次发生 亲和命中的 时间
   __be32   nat_ip;         // 小端存储。
   __u16    nat_port ;      // 小端存储。
   __u8     pad[2];
};


struct {
  __uint(type, BPF_MAP_TYPE_LRU_HASH);
  __type(key, struct mapkey_affinity  );
  __type(value, struct  mapvalue_affinity );
  __uint(max_entries, DEFAULT_MAX_EBPF_MAP_ENTRIES);
  __uint(pinning, 1); /* 这个配合 golang 中的 pinPath，完成 路径 pin */
  //__uint(map_flags, 0);
} map_affinity SEC(".maps");


//======================================= map ： 存储 nat 记录


struct mapkey_nat_record {
   __u64   socket_cookie;       //  bpf_get_socket_cookie(ctx);
   __be32  nat_ip;    //   小端存储。  nat 后的 ip
   __u16   nat_port;       // 小端存储。
   __u8    proto;
   __u8    pad;
};

struct mapvalue_nat_record {
	__be32 original_dest_ip;  // 小端存储。
	__u16  original_dest_port;       // 小端存储。
   __u8    pad[2];
};


struct {
  __uint(type, BPF_MAP_TYPE_LRU_HASH);
  __type(key, struct mapkey_nat_record );
  __type(value, struct mapvalue_nat_record );
  __uint(max_entries, DEFAULT_MAX_EBPF_MAP_ENTRIES);
  __uint(pinning, 1); /* 这个配合 golang 中的 pinPath，完成 路径 pin */
  //__uint(map_flags, 0);
} map_nat_record SEC(".maps");

#endif



//======================================= map ： 存储 event

// no available backend
#define FAILURE_CODE_AGENT__NO_BACKEND 	                     1
// there is available backend, but failed to find it in backend map
#define FAILURE_CODE_AGENT__FIND_BACKEND_FAILURE             2
// failed to find node entry ip map
#define FAILURE_CODE_AGENT__FIND_NODE_ENTRY_IP_MAP_FAILURE   3
// failed to update affinity map
#define FAILURE_CODE_SYSTEM__UPDATE_AFFINITY_MAP_FAILURE     100
// failed to update natRecord map
#define FAILURE_CODE_SYSTEM__UPDATE_NAT_RECORD_MAP_FAILURE   101



// BPF_MAP_TYPE_PERF_EVENT_ARRAY 类型的 数据结构体，在 golang 中需要自定定义 struct

struct event_value {
    __u64  cgroup_id ;

    __be64 nat_v6ip_high ;    // 小端存储。
    __be64 nat_v6ip_low ;    // 小端存储。
    __be64 original_dest_v6ip_high ;  /* 小端存储。 dest ip */
    __be64 original_dest_v6ip_low ;  /* 小端存储。 dest ip */

    __be32 nat_v4ip ;    // 小端存储。
    __be32 original_dest_v4ip ;  /* 小端存储。 dest ip */

	__be16 nat_port;   // 小端存储。
	__be16 original_dest_port;   // 小端存储。
	__u32  pid;

    __u32  svc_id ;
    __u8   is_ipv4 ; /* 0 for ipv6 data, 1 for ipv4 data */
    __u8   is_success ; /* 1 for success , 0 for failure */
    __u8   nat_type ;  /* NAT_TYPE_SERVICE (  lowest priority  ) ,NAT_TYPE_REDIRECT ,  NAT_TYPE_BALANCING ( highest priority )  */
    __u8   failure_code;

    __u8   nat_mode;    /* 用于标识发生了哪种 IP 地址的 nat， 参考 NAT_MODE_* 标志 */
    __u8   protocol;
    __u8   pad[6];
} ;

// BPF_MAP_TYPE_PERF_EVENT_ARRAY 中的 key 和 value 并不存放真正的 数据， key 用来存放 cpu 索引， values 存放指向 perf event buffer 的地址
struct {
	__uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY );
	__uint(key_size, sizeof(__u32));  /* BPF_MAP_TYPE_PERF_EVENT_ARRAY */
	__uint(value_size, sizeof(__u32));
	__uint(pinning, 1);  /* 这个配合 golang 中的 pinPath，完成 路径 pin */
	__uint(max_entries, 4096); /*默认会将max_entries设置为系统中cpu个数*/
} map_event  SEC(".maps");

// BPF_MAP_TYPE_RINGBUF
//struct {
//	__uint(type, BPF_MAP_TYPE_RINGBUF);
//	__uint(max_entries, 1 << 24);
//} map_event SEC(".maps");


//======================================= map ： 存储工作配置

#define INDEX_DEBUG_LEVEL   0
#define INDEX_ENABLE_IPV4   1
#define INDEX_ENABLE_IPV6   2

/* This a array map for configuration
    0 :  debug level ( 0:VERBOSE , 1:INFO, 2:ERROR)
    1 :  enable ipv4 ( 0: disabled , 1:enabled)
    2 :  enable ipv6 ( 0: disabled , 1:enabled)
*/
// https://docs.kernel.org/bpf/map_array.html
struct {
        __uint(type, BPF_MAP_TYPE_ARRAY); // All array elements are pre-allocated and zero initialized when created
        __type(key, __u32); // key 就是数组中的索引（index）（因此 key 一定 是整形），因此无需对 key 进行哈希 . All array elements are pre-allocated and zero initialized at init time. Key is an index in array and can only be 4 bytes (32-bit)
        __type(value, __u32);  // The value stored can be of any size , however, all array elements are aligned to 8 bytes
        __uint(pinning, 1);
        __uint(max_entries, 10);
} map_configure SEC(".maps");

#endif /* __MAP_H__ */
