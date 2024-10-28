package ebpf

import (
	"encoding/binary"
	"fmt"
	"net"
)

// ----------------- map flag ---------------------
var (
	NAT_TYPE_SERVICE   = uint8(0)
	NAT_TYPE_REDIRECT  = uint8(1)
	NAT_TYPE_BALANCING = uint8(2)
)
var NatTypeMap = map[uint8]string{
	NAT_TYPE_SERVICE:   "service",
	NAT_TYPE_REDIRECT:  "localRedirect",
	NAT_TYPE_BALANCING: "balancing",
}

var (
	IPPROTO_TCP = uint8(6)
	IPPROTO_UDP = uint8(17)
)
var ProtocalMap = map[uint8]string{
	IPPROTO_TCP: "tcp",
	IPPROTO_UDP: "udp",
}

// in line with NAT_MODE_* in the map.h
var (
	//
	NatModeServiceClusterip    = uint8(0x01)
	NatModeServiceLoadBalancer = uint8(0x02)
	NatModeServiceExternalIp   = uint8(0x03)
	NatModeServiceNodePort     = uint8(0x04)
	//
	NatModeRedirectService = uint8(0x10)
	NatModeRedirectAddress = uint8(0x11)
	//
	NatModeBalancingPod       = uint8(0x20)
	NatModeBalancingHostPort  = uint8(0x21)
	NatModeBalancingNodeProxy = uint8(0x22)
	NatModeBalancingAddress   = uint8(0x23)
)

var NatModeMap = map[uint8]string{
	NatModeServiceClusterip:    "ServiceClusterIP",
	NatModeServiceLoadBalancer: "ServiceLoadbalancer",
	NatModeServiceExternalIp:   "ServiceExternalIP",
	NatModeServiceNodePort:     "ServiceNodeport",
	//
	NatModeRedirectService: "localRedirectService",
	NatModeRedirectAddress: "localRedirectAddress",
	//
	NatModeBalancingPod:       "balancingPod",
	NatModeBalancingHostPort:  "balancingHostPort",
	NatModeBalancingNodeProxy: "balancingNodeProxy",
	NatModeBalancingAddress:   "balancingAddress",
}

var FailureCodeMap = map[uint8]string{
	0:   "",
	1:   "AgentNoBackend",
	2:   "AgentFindBackendFailure",
	3:   "AgentFindNodeProxyIpFailure",
	100: "SystemUpdateAffinityFailure",
	101: "SystemUpdateNatRecordFailure",
}

var (
	//
	SCOPE_LOCAL_CLUSTER = uint8(0)
	// for NodePorts, ExternalIPs, and LoadBalancer IPs
	SERVICE_FLAG_EXTERNAL_LOCAL_SVC = uint8(0x1)
	// for ClusterIP
	SERVICE_FLAG_INTERNAL_LOCAL_SVC = uint8(0x2)
	//
	NODEPORT_V4_IP = net.ParseIP("255.255.255.255").To4()
)

// -------------------------

func GetNatTypeStr(p uint8) string {
	if v, ok := NatTypeMap[p]; ok {
		return v
	} else {
		return ""
	}
}

func GetProtocolStr(p uint8) string {
	if v, ok := ProtocalMap[p]; ok {
		return v
	} else {
		return "unknown"
	}
}

func GetIpStr(p uint32) string {
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, p)
	return net.IP(tmp).String()
}

func GetNatModeStr(p uint8) string {
	if v, ok := NatModeMap[p]; ok {
		return v
	} else {
		return "unkown"
	}
}

func GetFailureStr(p uint8) string {
	if v, ok := FailureCodeMap[p]; ok {
		return v
	} else {
		return ""
	}
}

// -----------------------------------------------------

func (t bpf_cgroupMapkeyService) String() string {
	return fmt.Sprintf(`{ DestIp:%s, DestPort:%d, protocol:%s, NatType:%s, Scope:%d }`,
		GetIpStr(t.Address), t.Dport, GetProtocolStr(t.Proto), GetNatTypeStr(t.NatType), t.Scope)
}

func (t bpf_cgroupMapvalueService) String() string {
	return fmt.Sprintf(`{ SvcId:%d, TotalBackendCount:%d, LocalBackendCount:%d, AffinitySecond:%d, NatMode:%s, ServiceFlags:%d, BalancingFlags:%d, RedirectFlags:%d }`,
		t.SvcId, t.TotalBackendCount, t.LocalBackendCount, t.AffinitySecond, GetNatModeStr(t.NatMode), t.ServiceFlags, t.BalancingFlags, t.RedirectFlags)
}

// ------------------------------------------------

type bpf_cgroupMapkeyNodeIp struct {
	IpAddr uint32
}

func (t bpf_cgroupMapkeyNodeIp) String() string {
	return fmt.Sprintf(`{ NodeIp:%s}`, GetIpStr(t.IpAddr))
}

type bpf_cgroupMapvalueNodeProxyIp struct {
	IpAddr uint32
}

func (t bpf_cgroupMapvalueNodeProxyIp) String() string {
	return fmt.Sprintf(`{ NodeIp:%s}`, GetIpStr(t.IpAddr))
}

// ------------------------------------------------
func (t bpf_cgroupMapkeyBackend) String() string {
	return fmt.Sprintf(`{ Order:%d, SvcId:%d, port:%d, protocol:%s, NatType:%s, Scope: %d }`,
		t.Order, t.SvcId, t.Dport, GetProtocolStr(t.Proto), GetNatTypeStr(t.NatType), t.Scope)
}

func (t bpf_cgroupMapvalueBackend) String() string {
	return fmt.Sprintf(`{ PodIp:%s , PodPort:%d, NodeId:%d, NodePort:%d }`,
		GetIpStr(t.PodAddress), t.PodPort, t.NodeId, t.NodePort)
}

// ------------------------------------------------

func (t bpf_cgroupMapkeyNatRecord) String() string {
	return fmt.Sprintf(`{ SocketCookie:%d, NatIp:%s, NatPort:%d, protocol:%s }`,
		t.SocketCookie, GetIpStr(t.NatIp), t.NatPort, GetProtocolStr(t.Proto))
}

func (t bpf_cgroupMapvalueNatRecord) String() string {
	return fmt.Sprintf(`{ OriginalDstIp:%s , OriginalDstPort:%d }`,
		GetIpStr(t.OriginalDestIp), t.OriginalDestPort)
}

// --------------------------------------------------

func (t bpf_cgroupMapkeyAffinity) String() string {
	return fmt.Sprintf(`{ ClientCookie:%d , OriginalDestIp:%s, OriginalPort:%d, protocol:%s }`,
		t.ClientCookie, GetIpStr(t.OriginalDestIp), t.OriginalPort, GetProtocolStr(t.Proto))
}

func (t bpf_cgroupMapvalueAffinity) String() string {
	return fmt.Sprintf(`{ LastUpatedTimeStamp:%d , NatIp:%s, NatPort:%d  }`,
		t.Ts, GetIpStr(t.NatIp), t.NatPort)
}

// -------------------------------------------------

// struct for ebpf map : event
type MapEventValue struct {
	CgroupId             uint64
	NatV6ipHigh          uint64
	NatV6ipLow           uint64
	OriginalDestV6ipHigh uint64
	OriginalDestV6ipLow  uint64
	NatV4Ip              uint32
	OriginalDestV4Ip     uint32
	NatPort              uint16
	OriginalDestPort     uint16
	Pid                  uint32
	SvcId                uint32
	IsIpv4               uint8 /* 0 for ipv6 data, 1 for ipv4 data */
	IsSuccess            uint8 /* 1 for success , 0 for failure */
	NatType              uint8 /* 1 for NAT_TYPE_FLOATIP , 2 for NAT_TYPE_SVC, 3 for NAT_TYPE_REDIRECT  */
	FailureCode          uint8
	NatMode              uint8
	Protocol             uint8
	Pad                  [6]uint8
}

func GetIpv6Str(ipV6High, ipV6Low uint64) string {
	ip := make([]byte, 16)
	for i := 0; i < 8; i++ {
		ip[i] = byte(ipV6High >> (8 * (7 - i)))
		ip[i+8] = byte(ipV6Low >> (8 * (7 - i)))
	}
	return net.IP(ip).String()
}

func (t MapEventValue) String() string {
	return fmt.Sprintf(`{ CgroupId:%d, IsIpv4:%d, SvcId:%d, IsSuccess:%d, NatType:%s, NatMode:%s, Protocol:%s, OriginalDestV4Ip:%s, OriginalDestV6Ip:%s, OriginalDestPort:%d, NatV4Ip:%s, NatV6Ip:%s, NatPort:%d , Pid:%d, FailureCode:%s }`,
		t.CgroupId, t.IsIpv4, t.SvcId, t.IsSuccess, GetNatTypeStr(t.NatType), GetNatModeStr(t.NatMode), GetProtocolStr(t.Protocol),
		GetIpStr(t.OriginalDestV4Ip), GetIpv6Str(t.OriginalDestV6ipHigh, t.OriginalDestV6ipLow), t.OriginalDestPort, GetIpStr(t.NatV4Ip), GetIpv6Str(t.NatV6ipHigh, t.NatV6ipLow), t.NatPort, t.Pid, GetFailureStr(t.FailureCode))
}

// ------------------------------------------------- map configure

const (
	MapConfigureKeyIndexDebugLevel = iota
	MapConfigureKeyIndexIpv4Enabled
	MapConfigureKeyIndexIpv6Enabled
	MapConfigureKeyIndexEnd
)

const (
	// for IndexDebugLevel
	MapConfigureValueDebugLevelVerbose = 0
	MapConfigureValueDebugLevelInfo    = 1
	MapConfigureValueDebugLevelError   = 2
	// for IndexIpv4Enabled and Ipv6Enabled
	MapConfigureValueDisabled = 0
	MapConfigureValueEnabled  = 1
)

func MapConfigureStr(key, value uint32) string {
	switch key {
	case MapConfigureKeyIndexDebugLevel:
		if value == MapConfigureValueDebugLevelVerbose {
			return "debugLevel: versbose"
		} else if value == MapConfigureValueDebugLevelInfo {
			return "debugLevel: info"
		} else if value == MapConfigureValueDebugLevelError {
			return "debugLevel: error"
		} else {
			return "debugLevel: uknown"
		}
	case MapConfigureKeyIndexIpv4Enabled:
		if value == 0 {
			return "Ipv4Enabled: disabled"
		} else {
			return "Ipv4Enabled: enabled"
		}
	case MapConfigureKeyIndexIpv6Enabled:
		if value == 0 {
			return "Ipv6Enabled: disabled"
		} else {
			return "Ipv6Enabled: enabled"
		}
	}
	return "unknowKey: uknowValue"
}
