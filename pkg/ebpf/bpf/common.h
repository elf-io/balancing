#ifndef __COMMON_H__
#define __COMMON_H__


#include "vmlinux.h"
//#include <linux/ptrace.h>
//#include <stdio.h>
//#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>


//------------------

#define DEBUG_LEVEL DEBUG_VERSBOSE
#define ENABLE_IPV4


//-----------------

char __license[] SEC("license") = "Dual BSD/GPL";



// macro BPFTRACE_HAVE_BTF is defined if BTF is detected
//#ifndef BPFTRACE_HAVE_BTF
//#else
//#endif

//#define AF_INET   2 /* IPv4 */
//#define AF_INET6 10 /* IPv6 */



#if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
#define bpf_htons(x) __builtin_bswap16(x)
#define bpf_htonl(x) __builtin_bswap32(x)
#elif __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
#define bpf_htons(x) (x)
#define bpf_htonl(x) (x)
#else
#error "__BYTE_ORDER__ error"
#endif

#if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
#define bpf_htons(x) __builtin_bswap16(x)
#define bpf_htonl(x) __builtin_bswap32(x)
#elif __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
#define bpf_htons(x) (x)
#define bpf_htonl(x) (x)
#else
#error "__BYTE_ORDER__ error"
#endif


enum debug_level {
	DEBUG_VERSBOSE=0,
	DEBUG_INFO,
	DEBUG_ERROR,
};

#ifndef DEBUG_LEVEL
// do nothing
#define debugf(level, fmt, ...) ({})
#else
#define debugf(level, fmt, ...)                                     \
    ({                                                              \
        if( level >= DEBUG_LEVEL ) {                                \
            if ( level == DEBUG_VERSBOSE ) {                         \
                char ____fmt[] = "[verb] " fmt ;                  \
                bpf_trace_printk(____fmt, sizeof(____fmt), ##__VA_ARGS__);  \
            }else if  ( level == DEBUG_INFO ) {                         \
                char ____fmt[] = "[info] " fmt ;                  \
                bpf_trace_printk(____fmt, sizeof(____fmt), ##__VA_ARGS__);  \
            }else if  ( level == DEBUG_ERROR ) {                         \
                char ____fmt[] = "[error] " fmt ;                  \
                bpf_trace_printk(____fmt, sizeof(____fmt), ##__VA_ARGS__);  \
            }                                                           \
        }                                                               \
    })
#endif


//----------------------

#define SYS_REJECT	0
#define SYS_PROCEED	1



#define HOST_NETNS_COOKIE   bpf_get_netns_cookie(NULL)

typedef __u64 __net_cookie;
typedef __u64 __sock_cookie;


//--------------------------



//-----------------------





#endif /* __COMMON_H__ */
