package ebpfWriter

import (
	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
)

type balancingPolicyData struct {
	Policy  *balancingv1beta1.BalancingPolicy
	Svc     *corev1.Service
	Epslice *discovery.EndpointSlice
	// identical to the serviceId in the ebpf map, it is used for event to find policy
	// so only just update ServiceId before updating ebpf map
	ServiceId uint32
}
