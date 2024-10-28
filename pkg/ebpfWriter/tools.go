package ebpfWriter

import (
	"fmt"
	"net"
	"strings"

	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/podLabel"
	"github.com/elf-io/balancing/pkg/types"
	"github.com/elf-io/balancing/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ------------------------------------

func fakeEndpointSliceForRedirectPolicy(policy *balancingv1beta1.LocalRedirectPolicy) (*discovery.EndpointSlice, error) {
	eds := &discovery.EndpointSlice{}
	eds.Name = policy.Name
	eds.Namespace = types.NamespaceIgnore

	ipList := podLabel.PodLabelHandle.GetLocalIPWithLabelSelector(&policy.Spec.LocalRedirectBackend.LocalEndpointSelector)
	if len(ipList) == 0 {
		return nil, nil
	}

	for _, v := range ipList {
		t := discovery.Endpoint{}
		t.Addresses = []string{}
		if len(v.IPv4) > 0 {
			t.Addresses = append(t.Addresses, v.IPv4)
		}
		if len(v.IPv6) > 0 {
			t.Addresses = append(t.Addresses, v.IPv6)
		}
		t.NodeName = &v.NodeName // 确保使用正确的 NodeName
		eds.Endpoints = append(eds.Endpoints, t)
	}
	return eds, nil
}

// ------------------------------------

func (s *ebpfWriter) getNodeIp(nodeName string) (ipv4 string, ipv6 string) {
	s.ebpfNodeLock.Lock()
	node, ok := s.nodeData[nodeName]
	s.ebpfNodeLock.Unlock()
	if !ok {
		return
	}

	for _, v := range node.Status.Addresses {
		t := net.ParseIP(v.Address)
		if t == nil {
			continue
		}
		if t.To4() != nil {
			ipv4 = t.To4().String()
		} else {
			ipv6 = t.To16().String()
		}
	}
	return
}

func (s *ebpfWriter) getNodeProxyIp(nodeName string) (ipv4 string, ipv6 string) {
	s.ebpfNodeLock.Lock()
	node, ok := s.nodeData[nodeName]
	s.ebpfNodeLock.Unlock()
	if !ok {
		return
	}

	if entryIp, ok := node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv4]; ok {
		t := net.ParseIP(entryIp)
		if t != nil && t.To4() != nil { // 修正条件
			ipv4 = t.To4().String()
		}
	}

	if entryIp, ok := node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv6]; ok {
		t := net.ParseIP(entryIp)
		if t != nil && t.To4() == nil { // 修正条件
			ipv6 = t.To16().String()
		}
	}
	return
}

func (s *ebpfWriter) fakeEndpointSliceForBalancingPolicy(policy *balancingv1beta1.BalancingPolicy) (*discovery.EndpointSlice, error) {
	eds := &discovery.EndpointSlice{}
	eds.Name = policy.Name
	eds.Namespace = types.NamespaceIgnore

	if policy.Spec.BalancingBackend.ServiceEndpoint != nil {
		ipList := podLabel.PodLabelHandle.GetGlobalIPWithLabelSelector(&policy.Spec.BalancingBackend.ServiceEndpoint.EndpointSelector)
		if len(ipList) == 0 {
			return nil, nil
		}

		for _, v := range ipList {
			t := discovery.Endpoint{}
			t.Addresses = []string{}

			switch policy.Spec.BalancingBackend.ServiceEndpoint.RedirectMode {
			case balancingv1beta1.RedirectModePodEndpoint:
				if len(v.IPv4) > 0 {
					t.Addresses = append(t.Addresses, v.IPv4)
				}
				if len(v.IPv6) > 0 {
					t.Addresses = append(t.Addresses, v.IPv6)
				}
			case balancingv1beta1.RedirectModeHostPort:
				// for hostPort pod, only one pod could run in a same node, so no need to eliminate node IP redundancy
				ipv4, ipv6 := s.getNodeIp(v.NodeName)
				if len(ipv4) == 0 && len(ipv6) == 0 {
					s.log.Sugar().Errorf("failed to find node ip for node %s, so the pod %s/%s failed to locate its node IP for balancing policy %s", v.NodeName, v.IPv4, v.IPv6, policy.Name)
					continue
				}
				if len(ipv4) > 0 {
					t.Addresses = append(t.Addresses, ipv4)
				}
				if len(ipv6) > 0 {
					t.Addresses = append(t.Addresses, ipv6)
				}
			case balancingv1beta1.RedirectModeNodeProxy:
				ipv4, ipv6 := s.getNodeProxyIp(v.NodeName)
				if len(ipv4) == 0 && len(ipv6) == 0 {
					s.log.Sugar().Errorf("failed to find node proxy ip for node %s, so the pod %s/%s failed to locate its node proxy IP for balancing policy %s", v.NodeName, v.IPv4, v.IPv6, policy.Name)
					continue
				}
				if len(ipv4) > 0 {
					uniqe := true
				CHECK_IP_LOOP1:
					for _, c := range eds.Endpoints {
						for _, x := range c.Addresses {
							if strings.ToLower(x) == strings.ToLower(ipv4) {
								uniqe = false
								break CHECK_IP_LOOP1
							}
						}
					}
					if uniqe {
						t.Addresses = append(t.Addresses, ipv4)
					}
				}
				if len(ipv6) > 0 {
					uniqe := true
				CHECK_IP_LOOP2:
					for _, c := range eds.Endpoints {
						for _, x := range c.Addresses {
							if strings.ToLower(x) == strings.ToLower(ipv6) {
								uniqe = false
								break CHECK_IP_LOOP2
							}
						}
					}
					if uniqe {
						t.Addresses = append(t.Addresses, ipv6)
					}
				}
			default:
				return nil, fmt.Errorf("unknown redirect mode %v", policy.Spec.BalancingBackend.ServiceEndpoint.RedirectMode)
			}

			t.NodeName = &v.NodeName
			eds.Endpoints = append(eds.Endpoints, t)
		}
		return eds, nil
	}

	if policy.Spec.BalancingBackend.AddressEndpoint != nil {
		nodeName := types.NodeNameIgnore
		for _, v := range policy.Spec.BalancingBackend.AddressEndpoint.IPAddresses {
			t := discovery.Endpoint{}
			t.Addresses = []string{}
			t.Addresses = append(t.Addresses, v)
			t.NodeName = &nodeName
			eds.Endpoints = append(eds.Endpoints, t)
		}
		return eds, nil
	}

	return nil, fmt.Errorf("there is no BalancingBackend")
}

// ------------------------------------

func FakeServiceForRedirectPolicy(policy *balancingv1beta1.LocalRedirectPolicy) (*corev1.Service, error) {
	svc := &corev1.Service{}
	svc.Name = policy.Name
	svc.Namespace = types.NamespaceIgnore

	if utils.CheckIPv4Format(policy.Spec.RedirectFrontend.AddressMatcher.IP) {
		svc.Spec.IPFamilies = []corev1.IPFamily{corev1.IPv4Protocol}
	} else {
		svc.Spec.IPFamilies = []corev1.IPFamily{corev1.IPv6Protocol}
	}
	svc.Spec.ClusterIP = policy.Spec.RedirectFrontend.AddressMatcher.IP
	svc.Spec.ClusterIPs = []string{policy.Spec.RedirectFrontend.AddressMatcher.IP}
	svc.Spec.Type = corev1.ServiceTypeClusterIP

	if idStr, ok := policy.Annotations[types.AnnotationServiceID]; ok {
		svc.Annotations = make(map[string]string)
		svc.Annotations[types.AnnotationServiceID] = idStr
	} else {
		return nil, fmt.Errorf("failed to find annotation %s in the policy %s", types.AnnotationServiceID, policy.Name)
	}

	for _, v := range policy.Spec.RedirectFrontend.AddressMatcher.ToPorts {
		p, e := utils.StringToInt32(v.Port)
		if e != nil {
			return nil, fmt.Errorf("unsupported port %v for LocalRedirectPolicy %v: %v", v.Port, policy.Name, e)
		}
		t := corev1.ServicePort{
			Name: v.Name,
			Port: p,
		}
		if strings.Compare(strings.ToLower(v.Protocol), strings.ToLower(string(corev1.ProtocolTCP))) == 0 {
			t.Protocol = corev1.ProtocolTCP
		} else if strings.Compare(strings.ToLower(v.Protocol), strings.ToLower(string(corev1.ProtocolUDP))) == 0 {
			t.Protocol = corev1.ProtocolUDP
		} else {
			return nil, fmt.Errorf("unsupported protocol %v for LocalRedirectPolicy %v", v.Protocol, policy.Name)
		}
		ok := false
		for _, m := range policy.Spec.LocalRedirectBackend.ToPorts {
			if m.Name == v.Name && strings.Compare(strings.ToLower(m.Protocol), strings.ToLower(v.Protocol)) == 0 {
				p, e := utils.StringToInt32(m.Port)
				if e != nil {
					return nil, fmt.Errorf("unsupported port %v for LocalRedirectPolicy %v: %v", v.Port, policy.Name, e)
				}
				t.TargetPort = intstr.FromInt32(p)
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("did not find %v targetport for LocalRedirectPolicy %v", v.Name, policy.Name)
		}
		svc.Spec.Ports = append(svc.Spec.Ports, t)
	}
	return svc, nil
}

// ------------------------------------

func FakeServiceForBalancingPolicyByAddressMatcher(policy *balancingv1beta1.BalancingPolicy) (*corev1.Service, error) {
	if policy.Spec.BalancingFrontend.AddressMatcher == nil {
		return nil, fmt.Errorf("front AddressMatcher is nil ")
	}

	svc := &corev1.Service{}
	svc.Name = policy.Name
	svc.Namespace = types.NamespaceIgnore

	if utils.CheckIPv4Format(policy.Spec.BalancingFrontend.AddressMatcher.IP) {
		svc.Spec.IPFamilies = []corev1.IPFamily{corev1.IPv4Protocol}
	} else {
		svc.Spec.IPFamilies = []corev1.IPFamily{corev1.IPv6Protocol}
	}
	svc.Spec.ClusterIP = policy.Spec.BalancingFrontend.AddressMatcher.IP
	svc.Spec.ClusterIPs = []string{policy.Spec.BalancingFrontend.AddressMatcher.IP}
	svc.Spec.Type = corev1.ServiceTypeClusterIP

	if idStr, ok := policy.Annotations[types.AnnotationServiceID]; ok {
		svc.Annotations = make(map[string]string)
		svc.Annotations[types.AnnotationServiceID] = idStr
	} else {
		return nil, fmt.Errorf("failed to find annotation %s in the policy %s", types.AnnotationServiceID, policy.Name)
	}

	for _, v := range policy.Spec.BalancingFrontend.AddressMatcher.ToPorts {
		p, e := utils.StringToInt32(v.Port)
		if e != nil {
			return nil, fmt.Errorf("unsupported port %v for BalancingPolicy %v: %v", v.Port, policy.Name, e)
		}
		t := corev1.ServicePort{
			Name: v.Name,
			Port: p,
		}
		if strings.Compare(strings.ToLower(v.Protocol), strings.ToLower(string(corev1.ProtocolTCP))) == 0 {
			t.Protocol = corev1.ProtocolTCP
		} else if strings.Compare(strings.ToLower(v.Protocol), strings.ToLower(string(corev1.ProtocolUDP))) == 0 {
			t.Protocol = corev1.ProtocolUDP
		} else {
			return nil, fmt.Errorf("unsupported protocol %v for BalancingPolicy %v", v.Protocol, policy.Name)
		}
		ok := false
		if policy.Spec.BalancingBackend.ServiceEndpoint != nil {
			for _, m := range policy.Spec.BalancingBackend.ServiceEndpoint.ToPorts {
				if m.Name == v.Name && strings.Compare(strings.ToLower(m.Protocol), strings.ToLower(v.Protocol)) == 0 {
					p, e := utils.StringToInt32(m.Port)
					if e != nil {
						return nil, fmt.Errorf("unsupported port %v for BalancingPolicy %v: %v", m.Port, policy.Name, e)
					}
					t.TargetPort = intstr.FromInt32(p)
					ok = true
					break
				}
			}
		} else {
			for _, m := range policy.Spec.BalancingBackend.AddressEndpoint.ToPorts {
				if m.Name == v.Name && strings.Compare(strings.ToLower(m.Protocol), strings.ToLower(v.Protocol)) == 0 {
					p, e := utils.StringToInt32(m.Port)
					if e != nil {
						return nil, fmt.Errorf("unsupported port %v for BalancingPolicy %v: %v", v.Port, policy.Name, e)
					}
					t.TargetPort = intstr.FromInt32(p)
					ok = true
					break
				}
			}
		}
		if !ok {
			return nil, fmt.Errorf("did not find %v targetport for BalancingPolicy %v", v.Name, policy.Name)
		}
		svc.Spec.Ports = append(svc.Spec.Ports, t)
	}
	return svc, nil
}

func FakeServiceForBalancingPolicyByServiceMatcher(policy *balancingv1beta1.BalancingPolicy, oldSvc *corev1.Service) (svc *corev1.Service, err error) {
	svc = &corev1.Service{}

	if policy.Spec.BalancingFrontend.ServiceMatcher == nil {
		return nil, fmt.Errorf("front ServiceMatcher is nil ")
	}
	if oldSvc == nil {
		return nil, fmt.Errorf("Service is nil ")
	}

	svc.Name = oldSvc.Name
	svc.Namespace = oldSvc.Namespace
	svc.Spec.Type = corev1.ServiceTypeClusterIP
	svc.Spec.SessionAffinity = corev1.ServiceAffinityNone
	svc.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyCluster
	svc.Spec.InternalTrafficPolicy = nil
	svc.Status.LoadBalancer.Ingress = nil
	svc.Spec.IPFamilyPolicy = oldSvc.Spec.IPFamilyPolicy
	svc.Spec.ClusterIP = oldSvc.Spec.ClusterIP
	svc.Spec.ClusterIPs = []string{}
	for _, v := range oldSvc.Spec.ClusterIPs {
		svc.Spec.ClusterIPs = append(svc.Spec.ClusterIPs, v)
	}
	svc.Spec.IPFamilies = []corev1.IPFamily{}
	for _, v := range oldSvc.Spec.IPFamilies {
		svc.Spec.IPFamilies = append(svc.Spec.IPFamilies, v)
	}
	svc.Spec.Ports = []corev1.ServicePort{}
	svc.Annotations = map[string]string{
		types.AnnotationServiceID: policy.Annotations[types.AnnotationServiceID],
	}

LOOP_policyPort:
	for _, policyPort := range policy.Spec.BalancingFrontend.ServiceMatcher.ToPorts {
		port, e := utils.StringToInt32(policyPort.Port)
		if e != nil {
			return nil, fmt.Errorf("unsupported port %v for BalancingPolicy %v: %v", policyPort.Port, policy.Name, e)
		}
		for _, svcPort := range oldSvc.Spec.Ports {
			if strings.ToLower(string(svcPort.Protocol)) == strings.ToLower(policyPort.Protocol) && svcPort.Port == port {
				// succeeded to find the port
				newport := corev1.ServicePort{
					Name:       svcPort.Name,
					Protocol:   svcPort.Protocol,
					Port:       svcPort.Port,
					TargetPort: intstr.FromInt32(0),
				}
				findPort := false
				// find the TargetPort from the backend
				if policy.Spec.BalancingBackend.AddressEndpoint != nil {
					for _, toPort := range policy.Spec.BalancingBackend.AddressEndpoint.ToPorts {
						if toPort.Name == policyPort.Name && strings.ToLower(toPort.Protocol) == strings.ToLower(policyPort.Protocol) {
							port, e := utils.StringToInt32(toPort.Port)
							if e != nil {
								return nil, fmt.Errorf("unsupported backend port %v for BalancingPolicy %v: %v", toPort, policy.Name, e)
							}
							newport.TargetPort = intstr.FromInt32(port)
							findPort = true
							break
						}
					}
				} else {
					for _, toPort := range policy.Spec.BalancingBackend.ServiceEndpoint.ToPorts {
						if toPort.Name == policyPort.Name && strings.ToLower(toPort.Protocol) == strings.ToLower(policyPort.Protocol) {
							p, e := utils.StringToInt32(toPort.Port)
							if e != nil {
								return nil, fmt.Errorf("unsupported port %v for BalancingPolicy %v: %v", port, policy.Name, e)
							}
							newport.TargetPort = intstr.FromInt32(p)
							findPort = true
							break
						}
					}
				}
				if !findPort {
					return nil, fmt.Errorf("failed to find backend port %v for BalancingPolicy %v", policy.Name)
				}
				svc.Spec.Ports = append(svc.Spec.Ports, newport)
				continue LOOP_policyPort
			}
		}
		return nil, fmt.Errorf("failed to find policyPort %v in the service %s/%s", policyPort, oldSvc.Namespace, oldSvc.Name)
	}

	return svc, nil
}
