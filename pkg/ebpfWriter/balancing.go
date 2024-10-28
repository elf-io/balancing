// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package ebpfWriter

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	"reflect"

	"github.com/elf-io/balancing/pkg/ebpf"
	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/utils"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type balancingPolicyData struct {
	Policy *balancingv1beta1.BalancingPolicy
	// a faked service for writing ebpf
	Svc *corev1.Service
	// a faked EndpointSlice for writing ebpf
	Epslice *discovery.EndpointSlice
	// identical to the serviceId in the ebpf map, it is used for event to find policy
	// so only just update ServiceId before updating ebpf map
	ServiceId uint32
}

func (s *ebpfWriter) getBalancingNatMode(policy *balancingv1beta1.BalancingPolicy) *uint8 {
	if policy == nil {
		return nil
	}

	if policy.Spec.BalancingBackend.ServiceEndpoint != nil {
		switch policy.Spec.BalancingBackend.ServiceEndpoint.RedirectMode {
		case balancingv1beta1.RedirectModePodEndpoint:
			return &ebpf.NatModeBalancingPod
		case balancingv1beta1.RedirectModeHostPort:
			return &ebpf.NatModeBalancingHostPort
		case balancingv1beta1.RedirectModeNodeProxy:
			return &ebpf.NatModeBalancingNodeProxy
		default:
			s.log.Sugar().Errorf("unknown RedirectMode in policy %s", policy.Name)
			return nil
		}
	} else {
		return &ebpf.NatModeBalancingAddress
	}
	return nil
}

func (s *ebpfWriter) UpdateBalancingByPolicy(l *zap.Logger, policy *balancingv1beta1.BalancingPolicy) error {
	if policy == nil {
		return fmt.Errorf("empty policy")
	}

	l.Sugar().Debugf("Starting UpdateBalancingByPolicy for policy: %s", policy.Name)

	policy.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()}
	index := policy.Name

	s.balancingPolicyLock.Lock()
	defer s.balancingPolicyLock.Unlock()

	l.Sugar().Debugf("get lock")

	if _, ok := s.balancingPolicyData[index]; ok {
		l.Sugar().Debugf("only support policy creation, do not support modification, BalancingPolicy %s", index)
		return nil
	}

	frontReady, backReady := false, false
	policyData := &balancingPolicyData{Policy: policy}

	if policy.Spec.BalancingFrontend.ServiceMatcher != nil {
		l.Sugar().Debugf("deal with ServiceMatcher")

		t := policy.Spec.BalancingFrontend.ServiceMatcher
		index := t.Namespace + "/" + t.ServiceName
		var svc1, svc2 *corev1.Service
		var e error
		s.ebpfServiceLock.Lock()
		if svcData, ok := s.serviceData[index]; ok {
			if e := utils.DeepCopy(svcData.Svc, &svc1); e != nil {
				l.Sugar().Errorf("failed to DeepCopy service %s: %v", svcData.Svc.Name, e)
				s.ebpfServiceLock.Unlock()
				return fmt.Errorf("failed to DeepCopy service: %v", e)
			}
		} else {
			s.ebpfServiceLock.Unlock()
			l.Sugar().Errorf("did not find service %s for policy %v", index, policy.Name)
			goto PROCESS_EDS_LABEL
		}
		s.ebpfServiceLock.Unlock()

		l.Sugar().Debugf("FakeServiceForBalancingPolicyByServiceMatcher")
		svc2, e = FakeServiceForBalancingPolicyByServiceMatcher(policy, svc1)
		if e != nil {
			return fmt.Errorf("failed to FakeServiceForBalancingPolicyByServiceMatcher: %v", e)
		}
		if svc2 == nil {
			return fmt.Errorf("failed to FakeServiceForBalancingPolicyByServiceMatcher")
		}
		policyData.Svc = svc2
		frontReady = true
	} else {
		l.Sugar().Debugf("deal with AddressMatcher")

		if t, e := FakeServiceForBalancingPolicyByAddressMatcher(policy); e != nil {
			l.Sugar().Debugf("Failed to fake service for BalancingPolicy %s: %v", index, e)
			return e
		} else {
			policyData.Svc = t
			l.Sugar().Debugf("fake service for BalancingPolicy %s", index)
			frontReady = true
		}
	}

PROCESS_EDS_LABEL:
	if eds, e := s.fakeEndpointSliceForBalancingPolicy(policy); e != nil {
		l.Sugar().Errorf("Failed to fakeEndpointSliceForBalancingPolicy for BalancingPolicy %s: %v", index, e)
	} else if eds != nil && len(eds.Endpoints) > 0 {
		l.Sugar().Debugf("fakeEndpointSliceForBalancingPolicy")
		policyData.Epslice = eds
		backReady = true
	}

	s.balancingPolicyData[index] = policyData
	if backReady && frontReady {
		// update id
		w, ipv6flag := ebpf.GenerateSvcV4Id(policyData.Svc)
		if w == 0 {
			if ipv6flag {
				l.Sugar().Debug("ignore ipv6 service for BalancingPolicy policy")
			} else {
				l.Sugar().Errorf("failed to get serviceId for BalancingPolicy policy")
			}
			return fmt.Errorf("failed to get serviceId for BalancingPolicy policy")
		}
		policyData.ServiceId = w
		l.Sugar().Debugf("update ServiceId to %d", w)
		// update map
		t := map[string]*discovery.EndpointSlice{policyData.Epslice.Name: policyData.Epslice}
		if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_BALANCING, nil, policyData.Svc, nil, t, s.getBalancingNatMode(policy)); e != nil {
			l.Sugar().Errorf("Failed to write ebpf map for balancing policy %v: %v", index, e)
			return e
		}
		l.Sugar().Infof("Succeeded to UpdateEbpfMapForService for BalancingPolicy %s", index)
	}

	l.Sugar().Debugf("finish")

	return nil
}

func (s *ebpfWriter) DeleteBalancingByPolicy(l *zap.Logger, policyName string) error {
	if len(policyName) == 0 {
		return fmt.Errorf("empty policy")
	}
	l.Sugar().Debugf("Starting DeleteBalancingByPolicy for policy: %s", policyName)

	index := policyName
	s.balancingPolicyLock.Lock()
	defer s.balancingPolicyLock.Unlock()
	if d, ok := s.balancingPolicyData[index]; ok {
		t := map[string]*discovery.EndpointSlice{}
		if d.Epslice != nil && len(d.Epslice.Endpoints) > 0 {
			t[d.Epslice.Name] = d.Epslice
		}
		if e := s.ebpfhandler.DeleteEbpfMapForService(l, ebpf.NAT_TYPE_BALANCING, d.Svc, t, s.getBalancingNatMode(d.Policy)); e != nil {
			l.Sugar().Errorf("failed to delete ebpf map for balancing policy %v when policy is deleting: %v", index, e)
			return e
		}
		l.Sugar().Infof("succeeded to delete ebpf map for the balancingPolicy %s", index)
	}
	return nil
}

func (s *ebpfWriter) UpdateBalancingByService(l *zap.Logger, svc *corev1.Service) error {
	l.Sugar().Debugf("Starting UpdateBalancingByService for service: %s", svc.Name)

	s.balancingPolicyLock.Lock()
	defer s.balancingPolicyLock.Unlock()

	for policyName, data := range s.balancingPolicyData {
		if data.Policy.Spec.BalancingFrontend.ServiceMatcher != nil {
			if data.Policy.Spec.BalancingFrontend.ServiceMatcher.ServiceName == svc.Name && data.Policy.Spec.BalancingFrontend.ServiceMatcher.Namespace == svc.Namespace {
				frontChanged := false
				oldSvc := data.Svc
				if data.Svc == nil || !reflect.DeepEqual(data.Svc.Spec.ClusterIPs, svc.Spec.ClusterIPs) || !reflect.DeepEqual(data.Svc.Spec.Ports, svc.Spec.Ports) {
					// fake new service
					svcNew, e := FakeServiceForBalancingPolicyByServiceMatcher(data.Policy, svc)
					if e != nil {
						return fmt.Errorf("failed to FakeServiceForBalancingPolicyByServiceMatcher: %v", e)
					}
					if svcNew == nil {
						return fmt.Errorf("failed to FakeServiceForBalancingPolicyByServiceMatcher")
					}
					s.balancingPolicyData[policyName].Svc = svcNew
					frontChanged = true
					l.Sugar().Debugf("Service spec changed for policy: %s", policyName)
				}
				if frontChanged {
					// no need to update svcId here, because the svcId does not change for balancing policy
					t := map[string]*discovery.EndpointSlice{}
					if data.Epslice != nil && len(data.Epslice.Endpoints) > 0 {
						t[data.Epslice.Name] = data.Epslice
					}
					if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_BALANCING, oldSvc, s.balancingPolicyData[policyName].Svc, t, t, s.getBalancingNatMode(data.Policy)); e != nil {
						l.Sugar().Errorf("Failed to update ebpf map for balancing policy %v when service %s/%s is changing: %v", policyName, svc.Namespace, svc.Name, e)
					} else {
						l.Sugar().Infof("Succeeded to update ebpf map for balancing policy %v when service %s/%s is changing", policyName, svc.Namespace, svc.Name)
					}
				} else {
					l.Sugar().Debugf("Just update service for balancing policy %v when service %s/%s is changing", policyName, svc.Namespace, svc.Name)
				}
			}
		}
	}
	return nil
}

func (s *ebpfWriter) DeleteBalancingByService(l *zap.Logger, svc *corev1.Service) error {
	l.Sugar().Debugf("Starting DeleteBalancingByService for service: %s", svc.Name)

	s.balancingPolicyLock.Lock()
	defer s.balancingPolicyLock.Unlock()

	for policyName, data := range s.balancingPolicyData {
		if data.Policy.Spec.BalancingFrontend.ServiceMatcher != nil {
			if data.Svc != nil && data.Policy.Spec.BalancingFrontend.ServiceMatcher.ServiceName == svc.Name && data.Policy.Spec.BalancingFrontend.ServiceMatcher.Namespace == svc.Namespace {
				oldSvc := data.Svc
				s.balancingPolicyData[policyName].Svc = nil
				t := map[string]*discovery.EndpointSlice{}
				if data.Epslice != nil && len(data.Epslice.Endpoints) > 0 {
					t[data.Epslice.Name] = data.Epslice
				}
				// no need to update svcId here, because the svcId does not change for balancing policy
				if e := s.ebpfhandler.DeleteEbpfMapForService(l, ebpf.NAT_TYPE_BALANCING, oldSvc, t, s.getBalancingNatMode(data.Policy)); e != nil {
					l.Sugar().Errorf("Failed to delete ebpf map for balancing policy %v when service %s/%s is deleted: %v", policyName, svc.Namespace, svc.Name, e)
				} else {
					l.Sugar().Infof("succeeded to delete ebpf map for balancing policy %v when service %s/%s is deleted", policyName, svc.Namespace, svc.Name)
				}
			}
		}
	}
	return nil
}

// UpdateBalancingByPod updates the balancing policy when a pod changes.
// It checks if the pod matches any existing balancing policies and updates the eBPF map accordingly.
func (s *ebpfWriter) UpdateBalancingByPod(l *zap.Logger, pod *corev1.Pod) error {
	l.Sugar().Debugf("Starting UpdateBalancingByPod for pod: %s", pod.Name)

	s.balancingPolicyLock.Lock()
	defer s.balancingPolicyLock.Unlock()

	for policyName, data := range s.balancingPolicyData {
		if data.Policy.Spec.BalancingBackend.AddressEndpoint != nil {
			continue
		}

		labelSelector, err := metav1.LabelSelectorAsSelector(&data.Policy.Spec.BalancingBackend.ServiceEndpoint.EndpointSelector)
		if err != nil {
			l.Sugar().Errorf("failed to get LabelSelectorAsSelector for policy %s: %v", policyName, err)
			continue
		}
		if !labelSelector.Matches(labels.Set(pod.Labels)) {
			continue
		}
		l.Sugar().Debugf("influence BalancingPolicy %s when pod %s/%s changes", policyName, pod.Namespace, pod.Name)

		newEps, e := s.fakeEndpointSliceForBalancingPolicy(data.Policy)
		if e != nil {
			l.Sugar().Errorf("failed to fakeEndpointSliceForBalancingPolicy for BalancingPolicy %s when pod %s/%s changes: %v", policyName, pod.Namespace, pod.Name, e)
			continue
		}

		oldEpList := map[string]*discovery.EndpointSlice{}
		if data.Epslice != nil {
			oldEpList[data.Epslice.Name] = data.Epslice
			l.Sugar().Debugf("the changing pod %s/%s influences balancing policy %s, oldEndpoints: %v", pod.Namespace, pod.Name, policyName, data.Epslice.Endpoints)
		} else {
			l.Sugar().Debugf("the changing pod %s/%s influences balancing policy %s, oldEndpoints: nil", pod.Namespace, pod.Name, policyName)
		}

		newEpList := map[string]*discovery.EndpointSlice{}
		if newEps != nil && len(newEps.Endpoints) > 0 {
			newEpList[newEps.Name] = newEps
			l.Sugar().Debugf("the changing pod %s/%s influences balancing policy %s, newEndpoints: %v", pod.Namespace, pod.Name, policyName, newEps.Endpoints)
		} else {
			l.Sugar().Debugf("the changing pod %s/%s influences balancing policy %s, newEndpoints: nil", pod.Namespace, pod.Name, policyName)
		}

		if data.Svc == nil {
			l.Sugar().Infof("the service is nil, skip ")
			data.Epslice = newEps
			continue
		} else {
			// no need to update svcId here, because the svcId does not change for balancing policy
			if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_BALANCING, data.Svc, data.Svc, oldEpList, newEpList, s.getBalancingNatMode(data.Policy)); e != nil {
				l.Sugar().Errorf("failed to update ebpf map for balancing policy %v when pod %s/%s is changing: %v", policyName, pod.Namespace, pod.Name, e)
			} else {
				l.Sugar().Infof("succeeded to update ebpf map for balancing policy %v when pod %s/%s is changing ", policyName, pod.Namespace, pod.Name)
			}
			data.Epslice = newEps
		}
	}
	return nil
}

func (s *ebpfWriter) DeleteBalancingByPod(l *zap.Logger, pod *corev1.Pod) error {
	return s.UpdateBalancingByPod(l, pod)
}

// --------------------------

// when the nodeIP or nodeProxyIP changes , update the target IP
func (s *ebpfWriter) UpdateBalancingByNode(l *zap.Logger, node *corev1.Node) error {
	l.Sugar().Debugf("Starting UpdateBalancingByNode for node: %s", node.Name)

	s.balancingPolicyLock.Lock()
	defer s.balancingPolicyLock.Unlock()

	for policyName, data := range s.balancingPolicyData {
		if data.Policy.Spec.BalancingBackend.AddressEndpoint != nil {
			continue
		}

		newEps, e := s.fakeEndpointSliceForBalancingPolicy(data.Policy)
		if e != nil {
			l.Sugar().Errorf("failed to fakeEndpointSliceForBalancingPolicy for BalancingPolicy %s when node %s changes: %v", policyName, node.Name, e)
			continue
		}
		if len(newEps.Endpoints) == 0 {
			continue
		}

		// check whether this node influences this policy
		if reflect.DeepEqual(data.Epslice, newEps) {
			continue
		}

		// update
		oldEpList := map[string]*discovery.EndpointSlice{}
		if data.Epslice != nil {
			oldEpList[data.Epslice.Name] = data.Epslice
			l.Sugar().Debugf("the changing node %s influences balancing policy %s, oldEndpoints: %v", node.Name, policyName, data.Epslice.Endpoints)
		} else {
			l.Sugar().Debugf("the changing node %s influences balancing policy %s, oldEndpoints: nil", node.Name, policyName)
		}

		newEpList := map[string]*discovery.EndpointSlice{}
		if newEps != nil && len(newEps.Endpoints) > 0 {
			newEpList[newEps.Name] = newEps
			l.Sugar().Debugf("the changing node %s influences balancing policy %s, newEndpoints: %v", node.Name, policyName, newEps.Endpoints)
		} else {
			l.Sugar().Debugf("the changing node %s influences balancing policy %s, newEndpoints: nil", node.Name, policyName)
		}

		if data.Svc == nil {
			l.Sugar().Infof("the service is nil, skip ")
			data.Epslice = newEps
			continue
		} else {
			// no need to update svcId here, because the svcId does not change for balancing policy
			if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_BALANCING, data.Svc, data.Svc, oldEpList, newEpList, s.getBalancingNatMode(data.Policy)); e != nil {
				l.Sugar().Errorf("failed to update ebpf map for balancing policy %v when node %s is changing: %v", policyName, node.Name, e)
			} else {
				l.Sugar().Infof("succeeded to update ebpf map for balancing policy %v when node %s is changing ", policyName, node.Name)
			}
			data.Epslice = newEps
		}

	}
	return nil
}
