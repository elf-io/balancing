#include "common.h"
#include "map.h"

/* use BPF_MAP_TYPE_RINGBUF */
//static __always_inline void ringbuf_output(  struct event_value *data )
//{
//    struct event_value  *event;
//    // require kernel 5.8
//    event = bpf_ringbuf_reserve( &map_event, sizeof(struct event_value), 0);
//    if (!event) {
//        debugf(DEBUG_ERROR, "error out perf event "  );
//        return
//    }
//
//    // must use ringbuf mem to submit
//    event->is_ipv4 = data->is_ipv4 ;
//    event->is_success = data->is_success ;
//    event->nat_type = data->nat_type ;
//    event->original_dest_ip = data->original_dest_ip ;
//    event->original_dest_port = data->original_dest_port ;
//    event->nat_ip = data->nat_ip ;
//    event->nat_port = data->nat_port ;
//
//    bpf_ringbuf_submit(event, 0);
//}

static __always_inline void perf_event_output(void *ctx, void *data, __u64 size ,  __u32 debug_level )
{
    int ret = bpf_perf_event_output(ctx, &map_event, BPF_F_CURRENT_CPU, data, size );
    if (ret) {
        debugf( DEBUG_ERROR, "error bpf_perf_event_output "  );
    }
    return ;
}


static __always_inline  bool ctx_in_hostns(void *ctx )
{
	__net_cookie own_cookie = bpf_get_netns_cookie(ctx);
	return own_cookie == HOST_NETNS_COOKIE ;
}



//-------------------------------------

static __always_inline bool get_service( __be32 dest_ip, __u16 dst_port, __u8 ip_proto  , struct mapkey_service *svckey , struct mapvalue_service *svcval ,  __u32 debug_level ) {
    struct mapvalue_service *t ;

    // search NAT_TYPE_BALANCING
    svckey->nat_type = NAT_TYPE_BALANCING ;
    svckey->address = dest_ip ;
    svckey->dport = dst_port ;
    svckey->proto = ip_proto ;
    svckey->scope = 0 ;
    svckey->pad[0] = 0 ;
    svckey->pad[1] = 0 ;
    svckey->pad[2] = 0 ;
    //debugf(DEBUG_INFO, "search : %pI4:%d  proto=%d \n" , &dest_ip, dst_port , ip_proto );
    //debugf(DEBUG_INFO, "search : nat_type=%d    \n" , svckey->nat_type  );
    t = bpf_map_lookup_elem( &map_service , svckey);
    if (t) {
        debugf(DEBUG_INFO, "get NAT_TYPE_BALANCING record \n" );
        goto succeed;
    }

    // search NAT_TYPE_REDIRECT
    svckey->nat_type = NAT_TYPE_REDIRECT ;
    t = bpf_map_lookup_elem( &map_service , svckey);
    if (t){
        debugf(DEBUG_INFO, "get NAT_TYPE_REDIRECT record \n" );
        goto succeed;
    }

    // search NAT_TYPE_SERVICE
    svckey->nat_type = NAT_TYPE_SERVICE ;
    t = bpf_map_lookup_elem( &map_service , svckey);
    if (t){
        debugf(DEBUG_INFO, "get NAT_TYPE_SERVICE record, which is not nodePort \n" );
        goto succeed;
    }else{
        // try to search nodePort
        __ipv4_ip ip4_addr = (__ipv4_ip)dest_ip ;
        __u32 *nodeval = bpf_map_lookup_elem( &map_node_ip , &ip4_addr);
        if ( nodeval ) {
            debugf( DEBUG_VERSBOSE, "dest address is the ip of a node\n" );
            // it is node ip
            // search service for nodePort
            svckey->address = NODE_PORT_IP ;
            t = bpf_map_lookup_elem( &map_service , svckey);
            if (t) {
                // it is nodePort
                debugf(DEBUG_INFO, "get NAT_TYPE_SERVICE record, which is nodePort \n" );
                goto succeed;
            }
        }
    }
    return false;

succeed:
    svcval->svc_id = t->svc_id ;
    svcval->total_backend_count = t->total_backend_count ;
    svcval->local_backend_count = t->local_backend_count ;
    svcval->affinity_second = t->affinity_second ;
    svcval->service_flags = t->service_flags ;
    svcval->balancing_flags = t->balancing_flags ;
    svcval->redirect_flags = t->redirect_flags ;
    svcval->nat_mode = t->nat_mode ;
    return true ;
}

static __always_inline struct mapvalue_affinity* get_affinity_and_update( struct bpf_sock_addr *ctx , __u32 affinity_second , __u8 ip_proto , __u32 debug_level ) {

    if (affinity_second == 0 ) {
        return NULL ;
    }

    struct mapkey_affinity affinityKey = {
       .original_dest_ip =  ctx->user_ip4 ,
       .client_cookie =  bpf_get_netns_cookie(ctx) ,
       .original_port = (__u16)(bpf_htonl(ctx->user_port)>>16) ,
       .proto = ip_proto ,
       .pad = 0 ,
    };
    //debugf(DEBUG_VERSBOSE, "search affinityKey  original_dest_ip=%pI4 \n"  ,  &(affinityKey.original_dest_ip)  );
    //debugf(DEBUG_VERSBOSE, "search affinityKey  client_cookie=%d \n"  ,  affinityKey.client_cookie  );
    //debugf(DEBUG_VERSBOSE, "search affinityKey  original_port=%d \n"  ,  affinityKey.original_port  );
    //debugf(DEBUG_VERSBOSE, "search affinityKey  ip_proto=%d \n"  ,  affinityKey.proto  );

    struct mapvalue_affinity *affinityValue = bpf_map_lookup_elem( &map_affinity , &affinityKey);
    if (!affinityValue) {
        return NULL;
    }

    // check timeout
    __u64 now = bpf_ktime_get_ns();
    if ( ( now - affinityValue->ts ) <= ( affinity_second  * 1000000000ULL ) ) {
        // .......... 需要检测下之前的 endpoint 是否还存活？否则 亲和解析 导致 访问失败

        //
        affinityValue->ts = bpf_ktime_get_ns();
        if ( bpf_map_update_elem(&map_affinity, &affinityKey, affinityValue ,  BPF_ANY) ) {
            debugf(DEBUG_ERROR, "failed to update map_affinity" );
            return NULL ;
        }
        return affinityValue ;
    }else{
        debugf(DEBUG_VERSBOSE, "the affinity record has expired \n"   );
    }

    return NULL ;
}



//----------------------

static __always_inline int execute_nat(struct bpf_sock_addr *ctx , __u32 debug_level ) {

	__u32 dst_ip = ctx->user_ip4;
	// user_port is saved in network order, convert to host order
	__u16 dst_port = (__u16)(bpf_htonl(ctx->user_port)>>16);

    __be32 nat_ip ;
    __be16 nat_port ;
    struct mapkey_nat_record natkey;
    struct mapvalue_nat_record natvalue;

    struct  event_value evt = {
        .cgroup_id = bpf_get_current_cgroup_id(),
        .nat_v6ip_high = 0 ,
        .nat_v6ip_low = 0 ,
        .original_dest_v6ip_high = 0 ,
        .original_dest_v6ip_low = 0 ,
        .is_ipv4 = 1 ,
        .is_success = 0 ,
        .original_dest_v4ip = dst_ip ,
        .original_dest_port = dst_port ,
        .pid = (__u32) ( 0x00000000ffffffff & bpf_get_current_pid_tgid() ),
        .failure_code = 0 ,
        .pad = 0 ,
        .nat_mode = 0 ,
    } ;

    if( ctx_in_hostns(ctx) ) {
        debugf(DEBUG_VERSBOSE, " in hostnetwork for %pI4:%d\n" , &dst_ip  , dst_port   );
    }else{
        debugf(DEBUG_VERSBOSE, " in pod for %pI4:%d\n" , &dst_ip  , dst_port   );
    }

	__u8 ip_proto;
	switch (ctx->type) {
	case SOCK_STREAM:
		debugf(DEBUG_VERSBOSE,"SOCK_STREAM -> assuming TCP for %pI4:%d\n" , &dst_ip  , dst_port   );
		ip_proto = IPPROTO_TCP;
		break;
	case SOCK_DGRAM:
		debugf(DEBUG_VERSBOSE,"SOCK_STREAM -> assuming UDP for %pI4:%d\n" , &dst_ip  , dst_port   );
		ip_proto = IPPROTO_UDP;
		break;
	default:
		debugf(DEBUG_VERSBOSE,"Unknown socket type: %d for %pI4:%d\n", (int)ctx->type , &dst_ip  , dst_port  );
		return 1 ;
	}


    // ------------- find service value
    struct mapkey_service svckey;
    struct mapvalue_service svcval;
    if ( ! get_service( dst_ip , dst_port , ip_proto , &svckey, &svcval , debug_level ) ) {
        // these packets may be forwarding for non-service
        debugf(DEBUG_VERSBOSE, "did not find service value for %pI4:%d\n" , &dst_ip  , dst_port   );
        return 2;
    }
    debugf(DEBUG_INFO, "succeeded to find service value for %pI4:%d\n" , &dst_ip  , dst_port   );
    evt.nat_mode=svcval.nat_mode ;

    __u32 backend_count = svcval.total_backend_count;
    if ( svcval.service_flags & (SERVICE_FLAG_INTERNAL_LOCAL_SVC | SERVICE_FLAG_EXTERNAL_LOCAL_SVC)  ) {
        backend_count = svcval.local_backend_count ;
        debugf(DEBUG_INFO, "forward to local backend for %pI4:%d\n" , &dst_ip  , dst_port   );
    }
    if ( backend_count == 0 ) {
        debugf(DEBUG_INFO, "no backend for %pI4:%d\n" , &dst_ip  , dst_port   );
        evt.failure_code = FAILURE_CODE_AGENT__NO_BACKEND ;
        goto output_event;
    }

    //------------ check affinity history
    if ( svcval.affinity_second > 0 ) {
        debugf(DEBUG_VERSBOSE, "search affinity service for %pI4:%d\n" ,&dst_ip  , dst_port   );
        struct mapvalue_affinity *affinityValue = get_affinity_and_update(ctx, svcval.affinity_second , ip_proto , debug_level ) ;
        if (affinityValue) {
            // update
            debugf(DEBUG_INFO, "nat by sencondary affinity, for %pI4:%d\n" , &dst_ip  , dst_port   );
            nat_ip = affinityValue->nat_ip ;
            nat_port = affinityValue->nat_port  ;
            goto set_nat ;
        }
    }

    // ----------------- get backend
    // ?? 使用了变量在  % 后边，使用数字，就报错 "R1 invalid mem access 'scalar'" 。 就不会报错
    __u32 a = bpf_get_prandom_u32();
    a %= backend_count ;
    struct mapkey_backend backendKey = {
    	.order = a,
        .svc_id = svcval.svc_id ,
        .dport = svckey.dport ,
        .proto = svckey.proto ,
        .nat_type = svckey.nat_type ,
        .scope = svckey.scope,
    };
    struct mapvalue_backend *backendValue = bpf_map_lookup_elem( &map_backend , &backendKey);
    if (!backendValue) {
        debugf(DEBUG_ERROR, "failed to find backend for %pI4:%d\n" , &dst_ip  , dst_port   );
        evt.failure_code = FAILURE_CODE_AGENT__FIND_BACKEND_FAILURE ;
        goto output_event;
    }

    if ( svcval.affinity_second > 0 ) {
        nat_ip = backendValue->pod_address ;
        nat_port = backendValue->pod_port ;

        // create affinity item
        struct mapkey_affinity affinityKey = {
           .original_dest_ip =  dst_ip ,
           .client_cookie =  bpf_get_netns_cookie(ctx) ,
           .original_port = dst_port ,
           .proto = ip_proto ,
           .pad = 0 ,
        };
        struct mapvalue_affinity affinityValue = {
           .nat_ip = nat_ip ,
           .nat_port = nat_port ,
           .ts = bpf_ktime_get_ns() ,
        };

        //debugf(DEBUG_VERSBOSE, "update affinityKey  original_dest_ip=%pI4 \n"  ,  &(affinityKey.original_dest_ip)  );
        //debugf(DEBUG_VERSBOSE, "update affinityKey  client_cookie=%d \n"  ,  affinityKey.client_cookie  );
        //debugf(DEBUG_VERSBOSE, "update affinityKey  original_port=%d \n"  ,  affinityKey.original_port  );
        //debugf(DEBUG_VERSBOSE, "update affinityKey  ip_proto=%d \n"  ,  affinityKey.proto  );

        if ( bpf_map_update_elem(&map_affinity, &affinityKey, &affinityValue , BPF_ANY) ) {
            debugf(DEBUG_ERROR, "failed to create map_affinity for %pI4:%d\n" , &dst_ip  , dst_port   );
            evt.failure_code = FAILURE_CODE_SYSTEM__UPDATE_AFFINITY_MAP_FAILURE ;
            goto output_event;
        }
        debugf(DEBUG_VERSBOSE, "nat by first affinity, for %pI4:%d\n" , &dst_ip  , dst_port   );
    }else{
        if ( svckey.nat_type == NAT_TYPE_BALANCING ) {
            evt.nat_type = NAT_TYPE_BALANCING ;
            if ( svcval.balancing_flags & NAT_FLAG_ACCESS_NODEPORT_BALANCING ) {
                __u32 node_id = backendValue->node_id ;
                __u32 *node_ip = bpf_map_lookup_elem( &map_node_entry_ip , &node_id);
                if (!node_ip) {
                    debugf(DEBUG_ERROR, "failed to find node entry ip for %pI4:%d\n" , &dst_ip  , dst_port   );
                    evt.failure_code = FAILURE_CODE_AGENT__FIND_NODE_ENTRY_IP_MAP_FAILURE ;
                    goto output_event;
                }
                nat_ip = *node_ip ;
                nat_port = backendValue->node_port ;
            }else{
                nat_ip = backendValue->pod_address ;
                nat_port = backendValue->pod_port ;
            }
            debugf(DEBUG_VERSBOSE, "nat by balancing, for %pI4:%d\n" , &dst_ip  , dst_port   );
        }else if ( svckey.nat_type == NAT_TYPE_REDIRECT ) {
            evt.nat_type = NAT_TYPE_REDIRECT ;
            nat_ip = backendValue->pod_address ;
            nat_port = backendValue->pod_port ;
            debugf(DEBUG_VERSBOSE, "nat by redirect, for %pI4:%d\n" , &dst_ip  , dst_port   );
        }else {
            evt.nat_type = NAT_TYPE_SERVICE ;
            nat_ip = backendValue->pod_address ;
            nat_port = backendValue->pod_port ;
            debugf(DEBUG_VERSBOSE, "nat by service, for %pI4:%d\n" , &dst_ip  , dst_port   );
        }
    }

set_nat:
    natkey.socket_cookie = bpf_get_socket_cookie(ctx) ;
    natkey.nat_ip = nat_ip ;
    natkey.nat_port = nat_port ;
    natkey.proto = ip_proto ;
    natkey.pad = 0 ;
    natvalue.original_dest_ip = dst_ip ;
    natvalue.original_dest_port = dst_port;
    natvalue.pad[0] = 0;
    natvalue.pad[1] = 0;
    if ( bpf_map_update_elem(&map_nat_record, &natkey, &natvalue , BPF_ANY) ) {
        debugf(DEBUG_ERROR, "failed to update map_nat_record for %pI4:%d\n" , &dst_ip  , dst_port   );
        evt.failure_code = FAILURE_CODE_SYSTEM__UPDATE_NAT_RECORD_MAP_FAILURE ;
    }

    ctx->user_ip4 = nat_ip ;
    ctx->user_port = bpf_htonl(((__u32)nat_port) << 16); ;

    debugf(DEBUG_INFO, " DNAT from %pI4:%d  " , &dst_ip  , dst_port  );
    debugf(DEBUG_INFO, " DNAT to %pI4:%d \n" , &nat_ip , nat_port );

    evt.is_success = 1 ;
    evt.nat_v4ip = nat_ip ;
    evt.nat_port = nat_port ;

output_event:
    // ringbuf_output( &e );
    perf_event_output(ctx, &evt , sizeof(evt) , debug_level ) ;

    return 0 ;
}

//----------------------

static __always_inline int get_configure(__u32 *debug_level , __u32 *ipv4_enabled , __u32 *ipv6_enabled ) {
    __u32 map_index ;
    __u32 *ptr ;
    char fmt[] = "elb cgroup: failed to get configure " ;
    char fmt_err[] = "elb cgroup: error, ipv4 and ipv6 are disabled at the meaning time" ;

    map_index = INDEX_DEBUG_LEVEL ;
    ptr = bpf_map_lookup_elem(&configure_map, &map_index);
    if (!ptr)  {
         bpf_trace_printk(fmt, sizeof(fmt) );
         return 1;
    }
    *debug_level = *ptr;

    map_index=INDEX_ENABLE_IPV4;
    ptr = bpf_map_lookup_elem(&configure_map, &map_index);
    if (!ptr)  {
         bpf_trace_printk(fmt, sizeof(fmt) );
         return 1;
    }
    *ipv4_enabled = *ptr;

    map_index=INDEX_ENABLE_IPV6;
    ptr = bpf_map_lookup_elem(&configure_map, &map_index);
    if (!ptr)  {
         bpf_trace_printk(fmt, sizeof(fmt) );
         return 1;
    }
    *ipv6_enabled = *ptr;

    if ( *ipv4_enabled == 0 &&  *ipv6_enabled == 0 ) {
         bpf_trace_printk(fmt_err, sizeof(fmt_err) );
    }

    return 0;
}

//----------------------------------

SEC("cgroup/connect4")
int sock4_connect(struct bpf_sock_addr *ctx) {
	int err;

    __u32 debug_level ;
    __u32 ipv4_enabled ;
    __u32 ipv6_enabled ;
    if ( get_configure(&debug_level, &ipv4_enabled, &ipv6_enabled)!=0 ) {
        return SYS_PROCEED;
    }
    if ipv4_enabled == 0 {
        return SYS_PROCEED;
    }

    //debugf(DEBUG_VERSBOSE, "connect4: dst_ip=%pI4 dst_port=%d\n" ,&dst_ip, bpf_htons(dst_port) );

    // invalid bpf_context access off=40 size=4 (43 line(s) omitted)
    //debugf(DEBUG_VERSBOSE, "connect4: src_ip=%pI4  \n" , ctx->msg_src_ip4  );

    // for tcp and udp
	err = execute_nat(ctx, debug_level );

	return SYS_PROCEED;
}

SEC("cgroup/sendmsg4")
int sock4_sendmsg(struct bpf_sock_addr *ctx)
{
	int err;

    __u32 debug_level ;
    __u32 ipv4_enabled ;
    __u32 ipv6_enabled ;
    if ( get_configure(&debug_level, &ipv4_enabled, &ipv6_enabled)!=0 ) {
        return SYS_PROCEED;
    }
    if ipv4_enabled == 0 {
        return SYS_PROCEED;
    }

    //debugf(DEBUG_VERSBOSE , "sendmsg4: dst_ip=%pI4 dst_port=%d\n" ,&dst_ip, bpf_htons(dst_port) );

    // for UDP
	err = execute_nat(ctx, *debug_val_ptr , debug_level );

	return SYS_PROCEED;
}


SEC("cgroup/recvmsg4")
int sock4_recvmsg(struct bpf_sock_addr *ctx)
{
	int err;

    __u32 debug_level ;
    __u32 ipv4_enabled ;
    __u32 ipv6_enabled ;
    if ( get_configure(&debug_level, &ipv4_enabled, &ipv6_enabled)!=0 ) {
        return SYS_PROCEED;
    }
    if ipv4_enabled == 0 {
        return SYS_PROCEED;
    }

    //debugf(DEBUG_VERSBOSE, "recvmsg4: dst_ip=%pI4 dst_port=%d\n" ,&dst_ip, bpf_htons(dst_port) );

	return SYS_PROCEED;
}

SEC("cgroup/getpeername4")
int sock4_getpeername(struct bpf_sock_addr *ctx)
{
	int err;

    __u32 debug_level ;
    __u32 ipv4_enabled ;
    __u32 ipv6_enabled ;
    if ( get_configure(&debug_level, &ipv4_enabled, &ipv6_enabled)!=0 ) {
        return SYS_PROCEED;
    }
    if ipv4_enabled == 0 {
        return SYS_PROCEED;
    }

    //debugf(DEBUG_VERSBOSE , "getpeername4: dst_ip=%pI4 dst_port=%d\n" ,&dst_ip, bpf_htons(dst_port) );

	return SYS_PROCEED;
}
