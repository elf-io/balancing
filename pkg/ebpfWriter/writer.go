// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package ebpfWriter

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpf"
	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/lock"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

type EbpfWriter interface {
	CleanEbpfMapData() error

	// for service
	UpdateServiceByService(*zap.Logger, *corev1.Service, bool) error
	DeleteServiceByService(*zap.Logger, *corev1.Service) error
	UpdateServiceByEndpointSlice(*zap.Logger, *discovery.EndpointSlice, bool) error
	DeleteServiceByEndpointSlice(*zap.Logger, *discovery.EndpointSlice) error

	// for node
	UpdateNode(*zap.Logger, *corev1.Node, bool) error
	DeleteNode(*zap.Logger, *corev1.Node) error

	// for localRedirect
	DeleteRedirectByPod(*zap.Logger, *corev1.Pod) error
	UpdateRedirectByPod(*zap.Logger, *corev1.Pod) error
	DeleteRedirectByService(*zap.Logger, *corev1.Service) error
	UpdateRedirectByService(*zap.Logger, *corev1.Service) error
	DeleteRedirectByPolicy(*zap.Logger, string) error
	UpdateRedirectByPolicy(*zap.Logger, *balancingv1beta1.LocalRedirectPolicy) error

	// for balancing
	DeleteBalancingByPod(*zap.Logger, *corev1.Pod) error
	UpdateBalancingByPod(*zap.Logger, *corev1.Pod) error
	DeleteBalancingByService(*zap.Logger, *corev1.Service) error
	UpdateBalancingByService(*zap.Logger, *corev1.Service) error
	DeleteBalancingByPolicy(*zap.Logger, string) error
	UpdateBalancingByPolicy(*zap.Logger, *balancingv1beta1.BalancingPolicy) error
	UpdateBalancingByNode(*zap.Logger, *corev1.Node) error

	// for ebpf event to find the service and policy
	GetPolicyBySvcId(uint8, uint32) (string, string, error)
}

type ebpfWriter struct {
	client *kubernetes.Clientset

	// ---- for service
	ebpfServiceLock *lock.Mutex
	// index: namesapce/serviceName
	serviceData map[string]*SvcEndpointData

	// ---- for node
	ebpfNodeLock *lock.Mutex
	nodeData     map[string]*corev1.Node

	// ---- for redirect
	redirectPolicyLock *lock.Mutex
	redirectPolicyData map[string]*redirectPolicyData

	// ----- for balancing
	balancingPolicyLock *lock.Mutex
	balancingPolicyData map[string]*balancingPolicyData

	// use the creationTimestamp to record the last update time, and calculate the validityTime
	validityTime time.Duration
	log          *zap.Logger
	ebpfhandler  ebpf.EbpfProgram
}

var _ EbpfWriter = (*ebpfWriter)(nil)

func NewEbpfWriter(c *kubernetes.Clientset, ebpfhandler ebpf.EbpfProgram, validityTime time.Duration, l *zap.Logger) EbpfWriter {
	t := ebpfWriter{
		client:              c,
		ebpfServiceLock:     &lock.Mutex{},
		redirectPolicyLock:  &lock.Mutex{},
		ebpfNodeLock:        &lock.Mutex{},
		balancingPolicyLock: &lock.Mutex{},
		serviceData:         make(map[string]*SvcEndpointData),
		nodeData:            make(map[string]*corev1.Node),
		redirectPolicyData:  make(map[string]*redirectPolicyData),
		balancingPolicyData: make(map[string]*balancingPolicyData),
		validityTime:        validityTime,
		log:                 l,
		ebpfhandler:         ebpfhandler,
	}

	go t.DeamonGC()
	return &t
}

func (s *ebpfWriter) CleanEbpfMapData() error {
	// before informer, clean all map data to keep all data up to date
	s.log.Sugar().Infof("clean ebpf map backend when startup ")
	s.ebpfhandler.CleanMapBackend()
	s.log.Sugar().Infof("clean ebpf map service when startup ")
	s.ebpfhandler.CleanMapService()
	s.log.Sugar().Infof("clean ebpf map nodeIp when startup ")
	s.ebpfhandler.CleanMapNodeIp()
	s.log.Sugar().Infof("clean ebpf map nodeProxyIp when startup ")
	s.ebpfhandler.CleanMapNodeProxyIp()
	return nil
}

func (s *ebpfWriter) DeamonGC() {
	// todo: delete ebpf map data according the metadata.CreationTimestamp by the validityTime
	logger := s.log
	logger.Sugar().Infof("ebpfWriter DeamonGC begin to retrieve ebpf data with validityTime %s", s.validityTime.String())
	for {
		time.Sleep(time.Hour)
	}
}

func (s *ebpfWriter) GetPolicyBySvcId(natType uint8, svcId uint32) (namespace string, name string, err error) {
	err = nil

	switch natType {
	case ebpf.NAT_TYPE_SERVICE:
		s.ebpfServiceLock.Lock()
		s.ebpfServiceLock.Unlock()
		for _, val := range s.serviceData {
			if val.ServiceId == svcId {
				if val.Svc != nil {
					namespace = val.Svc.Namespace
					name = val.Svc.Name
				}
				break
			}
		}
		if len(name) == 0 {
			err = fmt.Errorf("did not find any data")
		}
	case ebpf.NAT_TYPE_REDIRECT:
		s.redirectPolicyLock.Lock()
		s.redirectPolicyLock.Unlock()
		for _, val := range s.redirectPolicyData {
			if val.ServiceId == svcId {
				if val.Policy != nil {
					name = val.Policy.Name
				}
				break
			}
		}
		if len(name) == 0 {
			err = fmt.Errorf("did not find any data")
		}
	case ebpf.NAT_TYPE_BALANCING:
		s.balancingPolicyLock.Lock()
		s.balancingPolicyLock.Unlock()
		for _, val := range s.balancingPolicyData {
			if val.ServiceId == svcId {
				if val.Policy != nil {
					name = val.Policy.Name
				}
				break
			}
		}
		if len(name) == 0 {
			err = fmt.Errorf("did not find any data")
		}
	default:
		err = fmt.Errorf("unknown natType %d", natType)
	}
	return
}
