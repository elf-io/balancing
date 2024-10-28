// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package ebpf

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/elf-io/balancing/pkg/types"
	"github.com/elf-io/balancing/pkg/utils"
	"golang.org/x/sys/unix"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"net"
	"os"
	"path"
	"strings"
)

func checkMount(mPath string, mountType string) (bool, error) {
	// ms, err := os.ReadFile(HostMountInfoPath)
	// if err != nil {
	// 	return false, fmt.Errorf("failed to read mount file: %v", err)
	// }
	// mss := strings.Split(string(ms), "\n")
	// for _, m := range mss {
	// 	if strings.Contains(m, fmt.Sprintf(" %s %s ", mountPath, mountType)) {
	// 		return true, nil
	// 	}
	// }
	// return false, nil

	// in host
	mountsPath := "/proc/mounts"
	if utils.FileExists(path.Join(types.HostProcMountDir, mountsPath)) {
		// in container
		mountsPath = path.Join(types.HostProcMountDir, mountsPath)
	}

	// fmt.Printf("mountsPath: %s ,  %s:%s\n", mountsPath, mPath, mountType)

	f, err := os.Open(mountsPath)
	if err != nil {
		return false, fmt.Errorf("failed to read mount file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// example fields: cgroup2 /sys/fs/cgroup/unified cgroup2 rw,nosuid,nodev,noexec,relatime 0 0
		mountLine := scanner.Text()
		// fmt.Printf("%s \n", mountLine)
		fields := strings.Split(mountLine, " ")
		if len(fields) >= 3 && fields[2] == mountType && strings.Compare(fields[1], mPath) == 0 {
			return true, nil
		}
	}
	return false, nil

}

func checkOrMountCgroupV2(cgroupRoot string) error {

	if mount, err := checkMount(cgroupRoot, "cgroup2"); err != nil {
		return fmt.Errorf("failed to checkMount: %v", err)
	} else {
		if mount {
			fmt.Printf("cgroupV2 %s is already mounted \n", cgroupRoot)
			return nil
		}
	}
	fmt.Printf("begin to mount cgroupV2 fs: %s \n", cgroupRoot)

	cgroupRootStat, err := os.Stat(cgroupRoot)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cgroupRoot, 0755); err != nil {
				return fmt.Errorf("Unable to create cgroup mount directory: %w", err)
			}
		} else {
			return fmt.Errorf("Failed to stat the mount path %s: %w", cgroupRoot, err)
		}
	} else if !cgroupRootStat.IsDir() {
		return fmt.Errorf("%s is a file which is not a directory", cgroupRoot)
	}

	if err := unix.Mount("none", cgroupRoot, "cgroup2", 0, ""); err != nil {
		return fmt.Errorf("failed to mount %s: %w", cgroupRoot, err)
	}
	fmt.Printf("succeeded mounting cgroupV2 fs on %s \n", cgroupRoot)

	return nil
}

func checkOrMountBpfFs(bpfPath string) error {

	if mount, err := checkMount(bpfPath, "bpf"); err != nil {
		return fmt.Errorf("failed to checkMount: %v", err)
	} else {
		if mount {
			fmt.Printf("bpf %s is already mounted \n", bpfPath)
			return nil
		}
	}
	fmt.Printf("begin to mount bpf fs: %s \n", bpfPath)

	var err error
	_, err = os.Stat(bpfPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(bpfPath, 0755); err != nil {
				return fmt.Errorf("unable to create bpf mount directory: %s", err)
			}
		}
	}

	err = unix.Mount(bpfPath, bpfPath, "bpf", 0, "")
	if err != nil {
		return fmt.Errorf("failed to mount %s: %s", bpfPath, err)
	}

	return nil
}

// --------------------------------------------------

// getClusterIPs returns ClusterIPs for given IPFamily associated with the service.
func getClusterIPs(svc *corev1.Service, ipFamily corev1.IPFamily) []net.IP {
	IPs := make([]net.IP, 0)
	if svc.Spec.ClusterIPs != nil {
		for _, addr := range svc.Spec.ClusterIPs {
			t := net.ParseIP(addr)
			if ipFamily == corev1.IPv4Protocol && t.To4() != nil {
				IPs = append(IPs, t)
			} else if ipFamily == corev1.IPv6Protocol && t.To4() == nil {
				IPs = append(IPs, t)
			}
		}
	}
	return IPs
}

func GenerateSvcV4Id(svc *corev1.Service) (id uint32, ipv6Flag bool) {
	ok := false
	for _, v := range svc.Spec.IPFamilies {
		if v == corev1.IPv4Protocol {
			ok = true
		}
	}
	if !ok {
		return 0, true
	}

	// for balancing and localRedirect policy, it uses the uniq annotation id
	if idStr, ok := svc.Annotations[types.AnnotationServiceID]; ok {
		if id, err := utils.StringToUint32(idStr); err == nil {
			return id, false
		}
	}

	// for service , it uses clusterIP
	// 使用 clusterip （假设 IP 地址唯一） 作为 service 之间的区别，它用于关联 一个 service 和 其所属的所有 endpoint
	t := net.ParseIP(svc.Spec.ClusterIP)
	if t.To4() != nil {
		return binary.LittleEndian.Uint32(t.To4()), false
	}
	for _, v := range svc.Spec.ClusterIPs {
		t := net.ParseIP(v)
		if t.To4() != nil {
			return binary.LittleEndian.Uint32(t.To4()), false
		}
	}
	return 0, false
}

func GetPortProtocol(svcPort *corev1.ServicePort) uint8 {
	if svcPort.Protocol == corev1.ProtocolTCP {
		return IPPROTO_TCP
	} else if svcPort.Protocol == corev1.ProtocolUDP {
		return IPPROTO_UDP
	} else {
		return 0
	}
}

func GetServiceAffinityTime(svc *corev1.Service) uint32 {
	if svc.Spec.SessionAffinity == corev1.ServiceAffinityClientIP {
		if svc.Spec.SessionAffinityConfig != nil && svc.Spec.SessionAffinityConfig.ClientIP != nil && svc.Spec.SessionAffinityConfig.ClientIP.TimeoutSeconds != nil {
			a := *(svc.Spec.SessionAffinityConfig.ClientIP.TimeoutSeconds)
			return uint32(a)
		}
	}
	return 0
}

// func GetTotalEndpointAccount(edsList map[string]*discovery.EndpointSlice) uint32 {
// 	num := uint32(0)
// 	for _, k := range edsList {
// 		if k != nil {
// 			num += uint32(len(k.Endpoints))
// 		}
// 	}
// 	return num
// }
//
// func GetLocalEndpointAccount(edsList map[string]*discovery.EndpointSlice) uint32 {
// 	num := uint32(0)
// 	for _, k := range edsList {
// 		if k != nil {
// 			for _, t := range k.Endpoints {
// 				if t.Hostname != nil && *t.Hostname == types.AgentConfig.LocalNodeName {
// 					num += 1
// 				}
// 			}
// 		}
// 	}
// 	return num
// }

func GetServiceFlag(svc *corev1.Service) uint8 {
	flag := uint8(0)
	if svc.Spec.ExternalTrafficPolicy == corev1.ServiceExternalTrafficPolicyLocal {
		flag = flag | SERVICE_FLAG_EXTERNAL_LOCAL_SVC
	}
	if svc.Spec.InternalTrafficPolicy != nil && *svc.Spec.InternalTrafficPolicy == corev1.ServiceInternalTrafficPolicyLocal {
		flag = flag | SERVICE_FLAG_INTERNAL_LOCAL_SVC
	}
	return flag
}

func GetEndpointIPv4Address(edp *discovery.Endpoint) uint32 {
	for _, k := range edp.Addresses {
		t := net.ParseIP(k)
		if t.To4() != nil {
			return binary.LittleEndian.Uint32(t.To4())
		}
	}
	return 0
}

func GetServiceV4LoadbalancerIP(svc *corev1.Service) []net.IP {
	r := []net.IP{}
	// .spec.loadBalancerIP field for a Service was deprecated in Kubernetes v1.24
	for _, v := range svc.Status.LoadBalancer.Ingress {
		t := net.ParseIP(v.IP).To4()
		if t != nil {
			r = append(r, t)
		}
	}
	return r
}

func ClassifyV4Endpoint(edsList map[string]*discovery.EndpointSlice) (localEp []*discovery.Endpoint, remoteEp []*discovery.Endpoint) {
	localEp = []*discovery.Endpoint{}
	remoteEp = []*discovery.Endpoint{}

	if len(edsList) == 0 {
		// 处理空映射的逻辑
	}
	checkV4Addr := func(j *discovery.Endpoint) bool {
		for _, k := range j.Addresses {
			if net.ParseIP(k).To4() != nil {
				return true
			}
		}
		return false
	}

	for _, k := range edsList {
		for _, v := range k.Endpoints {
			if v.NodeName != nil && *v.NodeName == types.AgentConfig.LocalNodeName {
				// check the validity of ipv4 address
				if checkV4Addr(&v) {
					localEp = append(localEp, &v)
				}
			} else {
				if checkV4Addr(&v) {
					remoteEp = append(remoteEp, &v)
				}
			}
		}
	}
	return localEp, remoteEp
}

func GetServiceV4AllVip(svc *corev1.Service) []net.IP {
	r := []net.IP{}
	r = append(r, getClusterIPs(svc, corev1.IPv4Protocol)...)
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		r = append(r, GetServiceV4LoadbalancerIP(svc)...)
	}
	if svc.Spec.ExternalIPs != nil {
		for _, v := range svc.Spec.ExternalIPs {
			t := net.ParseIP(v).To4()
			if net.ParseIP(v).To4() != nil {
				r = append(r, t)
			}
		}
	}
	return r
}
