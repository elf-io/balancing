// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package ebpfWriter

import (
	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/elf-io/balancing/pkg/lock"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"time"
)

type EbpfWriter interface {
	CleanEbpfMapData() error

	UpdateService(*zap.Logger, *corev1.Service, bool) error
	DeleteService(*zap.Logger, *corev1.Service) error

	UpdateEndpointSlice(*zap.Logger, *discovery.EndpointSlice, bool) error
	DeleteEndpointSlice(*zap.Logger, *discovery.EndpointSlice) error

	UpdateNode(*zap.Logger, *corev1.Node, bool) error
	DeleteNode(*zap.Logger, *corev1.Node) error
}

type EndpointData struct {
	Svc *corev1.Service
	// one endpointslice store 100 endpoints by default
	// index: namesapce/name
	EpsliceList map[string]*discovery.EndpointSlice
}

type ebpfWriter struct {
	// index: namesapce/name
	ebpfServiceLock *lock.Mutex
	endpointData    map[string]*EndpointData

	ebpfNodeLock *lock.Mutex
	nodeData     map[string]*corev1.Node

	// use the creationTimestamp to record the last update time, and calculate the validityTime
	validityTime time.Duration
	log          *zap.Logger
	ebpfhandler  ebpf.EbpfProgram
}

var _ EbpfWriter = (*ebpfWriter)(nil)

func NewEbpfWriter(ebpfhandler ebpf.EbpfProgram, validityTime time.Duration, l *zap.Logger) EbpfWriter {
	t := ebpfWriter{
		ebpfServiceLock: &lock.Mutex{},
		ebpfNodeLock:    &lock.Mutex{},
		endpointData:    make(map[string]*EndpointData),
		nodeData:        make(map[string]*corev1.Node),
		validityTime:    validityTime,
		log:             l,
		ebpfhandler:     ebpfhandler,
	}

	go t.DeamonGC()
	return &t
}

func (s *ebpfWriter) CleanEbpfMapData() error {
	// before informer, clean all map data to keep all data up to date
	s.log.Sugar().Infof("clean ebpf map backend when stratup ")
	s.ebpfhandler.CleanMapBackend()
	s.log.Sugar().Infof("clean ebpf map service when stratup ")
	s.ebpfhandler.CleanMapService()
	s.log.Sugar().Infof("clean ebpf map nodeIp when stratup ")
	s.ebpfhandler.CleanMapNodeIp()
	s.log.Sugar().Infof("clean ebpf map nodeProxyIp when stratup ")
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
