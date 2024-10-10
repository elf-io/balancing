// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package ebpfWriter

import (
	"fmt"
	"reflect"
	"time"

	"github.com/elf-io/balancing/pkg/ebpf"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/labels"

	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type redirectPolicyData struct {
	Policy  *balancingv1beta1.LocalRedirectPolicy
	Svc     *corev1.Service
	Epslice *discovery.EndpointSlice
}

// UpdateRedirectByPolicy updates the redirect policy based on the given LocalRedirectPolicy.
// It checks if the policy already exists and only supports creation, not modification.
// If the policy is new, it generates the necessary service and endpoint slice data
// and updates the eBPF map accordingly.
func (s *ebpfWriter) UpdateRedirectByPolicy(l *zap.Logger, policy *balancingv1beta1.LocalRedirectPolicy) error {
	if policy == nil {
		return fmt.Errorf("empty policy")
	}

	l.Sugar().Debugf("Starting UpdateRedirectByPolicy for policy: %s", policy.Name)

	policy.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()}
	index := policy.Name
	l.Sugar().Debugf("update the RedirectPolicy %s", index)

	s.redirectPolicyLock.Lock()
	defer s.redirectPolicyLock.Unlock()

	if _, ok := s.redirectPolicyData[index]; ok {
		l.Sugar().Debugf("only support policy creation, do not support modification, RedirectPolicy %s", index)
		return nil
	}

	frontReady, backReady := false, false
	policyData := &redirectPolicyData{Policy: policy}

	if policy.Spec.RedirectFrontend.ServiceMatcher != nil {
		t := policy.Spec.RedirectFrontend.ServiceMatcher
		index := t.Namespace + "/" + t.ServiceName
		s.ebpfServiceLock.Lock()
		if svcData, ok := s.serviceData[index]; ok {
			policyData.Svc = svcData.Svc
			frontReady = true
		}
		s.ebpfServiceLock.Unlock()
	} else {
		if t, e := fakeServiceByAddressMatcher(policy); e != nil {
			l.Sugar().Debugf("Failed to fake service for RedirectPolicy %s: %v", index, e)
			return e
		} else {
			policyData.Svc = t
			frontReady = true
		}
	}

	if eds, e := fakeEndpointSlice(policy); e != nil {
		l.Sugar().Debugf("Failed to fakeEndpointSlice for RedirectPolicy %s: %v", index, e)
	} else if eds != nil && len(eds.Endpoints) > 0 {
		policyData.Epslice = eds
		backReady = true
	}

	s.redirectPolicyData[index] = policyData
	if backReady && frontReady {
		t := map[string]*discovery.EndpointSlice{policyData.Epslice.Name: policyData.Epslice}
		if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_REDIRECT, nil, policyData.Svc, nil, t); e != nil {
			l.Sugar().Errorf("Failed to write ebpf map for redirect policy %v: %v", index, e)
			return e
		}
		l.Sugar().Infof("Succeeded to UpdateEbpfMapForService for RedirectPolicy %s", index)
	}
	return nil
}

// DeleteRedirectByPolicy deletes the redirect policy based on the given LocalRedirectPolicy.
// It removes the associated service and endpoint slice data from the eBPF map.
func (s *ebpfWriter) DeleteRedirectByPolicy(l *zap.Logger, policyName string) error {
	if len(policyName) == 0 {
		return fmt.Errorf("empty policy")
	}

	l.Sugar().Debugf("Starting DeleteRedirectByPolicy for policy: %s", policyName)

	index := policyName
	s.redirectPolicyLock.Lock()
	defer s.redirectPolicyLock.Unlock()

	if d, ok := s.redirectPolicyData[index]; ok {
		t := map[string]*discovery.EndpointSlice{}
		if d.Epslice != nil && len(d.Epslice.Endpoints) > 0 {
			t[d.Epslice.Name] = d.Epslice
		}
		if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_REDIRECT, d.Svc, nil, t, nil); e != nil {
			l.Sugar().Errorf("failed to delete ebpf map for redirect policy %v when policy is deleting: %v", index, e)
			return e
		}
		l.Sugar().Infof("succeeded to delete ebpf map for the RedirectPolicy %s", index)
	}
	return nil
}

// UpdateRedirectByService updates the redirect policy when a service changes.
// It checks if the service matches any existing redirect policies and updates the eBPF map accordingly.
func (s *ebpfWriter) UpdateRedirectByService(l *zap.Logger, svc *corev1.Service) error {
	l.Sugar().Debugf("Starting UpdateRedirectByService for service: %s", svc.Name)

	s.redirectPolicyLock.Lock()
	defer s.redirectPolicyLock.Unlock()

	for policyName, data := range s.redirectPolicyData {
		if data.Policy.Spec.RedirectFrontend.ServiceMatcher != nil {
			if data.Policy.Spec.RedirectFrontend.ServiceMatcher.ServiceName == svc.Name && data.Policy.Spec.RedirectFrontend.ServiceMatcher.Namespace == svc.Namespace {
				frontChanged := false
				oldSvc := data.Svc
				if data.Svc == nil || !reflect.DeepEqual(data.Svc.Spec, svc.Spec) {
					s.redirectPolicyData[policyName].Svc = svc
					frontChanged = true
					l.Sugar().Debugf("Service spec changed for policy: %s", policyName)
				}
				if frontChanged {
					t := map[string]*discovery.EndpointSlice{}
					if data.Epslice != nil && len(data.Epslice.Endpoints) > 0 {
						t[data.Epslice.Name] = data.Epslice
					}
					if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_REDIRECT, oldSvc, svc, t, t); e != nil {
						l.Sugar().Errorf("Failed to update ebpf map for redirect policy %v when service %s/%s is changing: %v", policyName, svc.Namespace, svc.Name, e)
					} else {
						l.Sugar().Infof("Succeeded to update ebpf map for redirect policy %v when service %s/%s is changing", policyName, svc.Namespace, svc.Name)
					}
				} else {
					l.Sugar().Debugf("Just update service for redirect policy %v when service %s/%s is changing", policyName, svc.Namespace, svc.Name)
				}
			}
		}
	}
	return nil
}

// DeleteRedirectByService deletes the redirect policy when a service is deleted.
// It removes the associated service data from the eBPF map.
func (s *ebpfWriter) DeleteRedirectByService(l *zap.Logger, svc *corev1.Service) error {
	l.Sugar().Debugf("Starting DeleteRedirectByService for service: %s", svc.Name)

	s.redirectPolicyLock.Lock()
	defer s.redirectPolicyLock.Unlock()

	for policyName, data := range s.redirectPolicyData {
		if data.Policy.Spec.RedirectFrontend.ServiceMatcher != nil {
			if data.Policy.Spec.RedirectFrontend.ServiceMatcher.ServiceName == svc.Name && data.Policy.Spec.RedirectFrontend.ServiceMatcher.Namespace == svc.Namespace {
				oldSvc := data.Svc
				s.redirectPolicyData[policyName].Svc = nil
				t := map[string]*discovery.EndpointSlice{}
				if data.Epslice != nil && len(data.Epslice.Endpoints) > 0 {
					t[data.Epslice.Name] = data.Epslice
				}
				if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_REDIRECT, oldSvc, nil, t, t); e != nil {
					l.Sugar().Errorf("Failed to delete ebpf map for redirect policy %v when service %s/%s is deleted: %v", policyName, svc.Namespace, svc.Name, e)
				} else {
					l.Sugar().Infof("succeeded to delete ebpf map for redirect policy %v when service %s/%s is deleted", policyName, svc.Namespace, svc.Name)
				}
			}
		}
	}
	return nil
}

// UpdateRedirectByPod updates the redirect policy when a pod changes.
// It checks if the pod matches any existing redirect policies and updates the eBPF map accordingly.
func (s *ebpfWriter) UpdateRedirectByPod(l *zap.Logger, pod *corev1.Pod) error {
	l.Sugar().Debugf("Starting UpdateRedirectByPod for pod: %s", pod.Name)

	s.redirectPolicyLock.Lock()
	defer s.redirectPolicyLock.Unlock()

	for policyName, data := range s.redirectPolicyData {
		labelSelector, err := metav1.LabelSelectorAsSelector(&data.Policy.Spec.LocalRedirectBackend.LocalEndpointSelector)
		if err != nil {
			l.Sugar().Errorf("failed to get LabelSelectorAsSelector for policy %s: %v", policyName, err)
			continue
		}
		if !labelSelector.Matches(labels.Set(pod.Labels)) {
			continue
		}
		l.Sugar().Debugf("influence RedirectPolicy %s when pod %s/%s changes", policyName, pod.Namespace, pod.Name)

		newEps, e := fakeEndpointSlice(data.Policy)
		if e != nil {
			l.Sugar().Errorf("failed to fakeEndpointSlice for RedirectPolicy %s when pod %s/%s changes: %v", policyName, pod.Namespace, pod.Name, e)
			continue
		}

		oldEpList := map[string]*discovery.EndpointSlice{}
		if data.Epslice != nil {
			oldEpList[data.Epslice.Name] = data.Epslice
			l.Sugar().Debugf("the changing pod %s/%s influences redirect policy %s, oldEndpoints: %v", pod.Namespace, pod.Name, policyName, data.Epslice.Endpoints)
		} else {
			l.Sugar().Debugf("the changing pod %s/%s influences redirect policy %s, oldEndpoints: nil", pod.Namespace, pod.Name, policyName)
		}

		newEpList := map[string]*discovery.EndpointSlice{}
		if newEps != nil && len(newEps.Endpoints) > 0 {
			newEpList[newEps.Name] = newEps
			l.Sugar().Debugf("the changing pod %s/%s influences redirect policy %s, newEndpoints: %v", pod.Namespace, pod.Name, policyName, newEps.Endpoints)
		} else {
			l.Sugar().Debugf("the changing pod %s/%s influences redirect policy %s, newEndpoints: nil", pod.Namespace, pod.Name, policyName)
		}

		if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_REDIRECT, data.Svc, data.Svc, oldEpList, newEpList); e != nil {
			l.Sugar().Errorf("failed to update ebpf map for redirect policy %v when pod %s/%s is changing: %v", policyName, pod.Namespace, pod.Name, e)
		} else {
			l.Sugar().Infof("succeeded to update ebpf map for redirect policy %v when pod %s/%s is changing ", policyName, pod.Namespace, pod.Name)
		}
		data.Epslice = newEps
	}
	return nil
}

// DeleteRedirectByPod deletes the redirect policy when a pod is deleted.
// It calls UpdateRedirectByPod to handle the deletion logic.
func (s *ebpfWriter) DeleteRedirectByPod(l *zap.Logger, pod *corev1.Pod) error {
	return s.UpdateRedirectByPod(l, pod)
}
