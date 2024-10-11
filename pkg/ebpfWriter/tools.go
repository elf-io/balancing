package ebpfWriter

import (
	"fmt"
	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/podLabel"
	"github.com/elf-io/balancing/pkg/types"
	"github.com/elf-io/balancing/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

func fakeEndpointSlice(policy *balancingv1beta1.LocalRedirectPolicy) (*discovery.EndpointSlice, error) {
	eds := &discovery.EndpointSlice{}
	eds.Name = policy.Name
	eds.Namespace = "faked"

	ipList := podLabel.PodLabelHandle.GetIPWithLabelSelector(&policy.Spec.LocalRedirectBackend.LocalEndpointSelector)
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
		t.NodeName = &types.AgentConfig.LocalNodeName
		eds.Endpoints = append(eds.Endpoints, t)
	}
	return eds, nil
}

func FakeServiceByAddressMatcher(policy *balancingv1beta1.LocalRedirectPolicy) (*corev1.Service, error) {
	svc := &corev1.Service{}
	svc.Name = policy.Name
	svc.Namespace = "faked"
	if utils.CheckIPv4Format(policy.Spec.RedirectFrontend.AddressMatcher.IP) {
		svc.Spec.IPFamilies = []corev1.IPFamily{corev1.IPv4Protocol}
	} else {
		svc.Spec.IPFamilies = []corev1.IPFamily{corev1.IPv6Protocol}
	}
	svc.Spec.ClusterIP = policy.Spec.RedirectFrontend.AddressMatcher.IP
	svc.Spec.ClusterIPs = []string{policy.Spec.RedirectFrontend.AddressMatcher.IP}
	svc.Spec.Type = corev1.ServiceTypeClusterIP
	svc.Annotations[types.AnnotationServiceID] = policy.Annotations[types.AnnotationServiceID]

	for _, v := range policy.Spec.RedirectFrontend.AddressMatcher.ToPorts {
		p, e := utils.StringToInt32(v.Port)
		if e != nil {
			return nil, fmt.Errorf("unspported port %v for LocalRedirectPolicy %v: %v", v.Port, policy.Name, e)
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
			return nil, fmt.Errorf("unspported protocol %v for LocalRedirectPolicy %v", v.Protocol, policy.Name)
		}
		ok := false
		for _, m := range policy.Spec.LocalRedirectBackend.ToPorts {
			if m.Name == v.Name && strings.Compare(strings.ToLower(m.Protocol), strings.ToLower(v.Protocol)) == 0 {
				p, e := utils.StringToInt32(m.Port)
				if e != nil {
					return nil, fmt.Errorf("unspported port %v for LocalRedirectPolicy %v: %v", v.Port, policy.Name, e)
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
