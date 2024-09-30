// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package ebpfWriter

import (
	"fmt"
	"github.com/elf-io/elf/pkg/k8s"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func shallowCopyEdpSliceMap(t map[string]*discovery.EndpointSlice) map[string]*discovery.EndpointSlice {
	m := make(map[string]*discovery.EndpointSlice)
	for k, v := range t {
		m[k] = v
	}
	return m
}

func (s *ebpfWriter) UpdateService(l *zap.Logger, svc *corev1.Service, onlyUpdateTime bool) error {

	if svc == nil {
		return fmt.Errorf("empty service")
	}

	// use it to record last update time
	svc.ObjectMeta.CreationTimestamp = metav1.Time{
		time.Now(),
	}

	index := svc.Namespace + "/" + svc.Name
	l.Sugar().Debugf("update the service %s", index)

	s.ebpfServiceLock.Lock()
	defer s.ebpfServiceLock.Unlock()
	if d, ok := s.endpointData[index]; ok {
		if d.EpsliceList != nil && len(d.EpsliceList) > 0 {
			if !onlyUpdateTime {
				l.Sugar().Infof("cache the data, and apply new data to ebpf map for service %v", index)
				if e := s.ebpfhandler.UpdateEbpfMapForService(l, d.Svc, svc, d.EpsliceList, d.EpsliceList); e != nil {
					l.Sugar().Errorf("failed to write ebpf map: %v", e)
					return e
				}
				d.Svc = svc
			} else {
				l.Sugar().Debugf("just update lastUpdateTime")
				d.Svc = svc
			}
		} else {
			l.Sugar().Debugf("cache the data, but no need to apply new data to ebpf map, cause miss endpointslice")
			d.Svc = svc
		}
	} else {
		l.Sugar().Debugf("cache the data, but no need to apply new data to ebpf map, cause miss endpointslice")
		s.endpointData[index] = &EndpointData{
			Svc:         svc,
			EpsliceList: make(map[string]*discovery.EndpointSlice),
		}
	}

	return nil
}

func (s *ebpfWriter) DeleteService(l *zap.Logger, svc *corev1.Service) error {
	if svc == nil {
		return fmt.Errorf("empty service")
	}

	index := svc.Namespace + "/" + svc.Name
	l.Sugar().Debugf("delete service %s", index)

	s.ebpfServiceLock.Lock()
	defer s.ebpfServiceLock.Unlock()
	if d, ok := s.endpointData[index]; ok {
		// todo : generate a ebpf map data and apply it
		l.Sugar().Infof("delete data from ebpf map for service: %v", index)
		if e := s.ebpfhandler.DeleteEbpfMapForService(l, d.Svc, d.EpsliceList); e != nil {
			l.Sugar().Errorf("failed to write ebpf map: %v", e)
			return e
		}
		delete(s.endpointData, index)
	} else {
		l.Sugar().Debugf("no need to delete service from ebpf map, cause already removed")
	}

	return nil
}

// -------------------------------------------------------------
func (s *ebpfWriter) UpdateEndpointSlice(l *zap.Logger, epSlice *discovery.EndpointSlice, onlyUpdateTime bool) error {

	if epSlice == nil {
		return fmt.Errorf("empty EndpointSlice")
	}
	epSlice.ObjectMeta.CreationTimestamp = metav1.Time{
		time.Now(),
	}

	// for default/kubernetes ，there is no owner
	index := k8s.GetEndpointSliceOwnerName(epSlice)
	epindex := epSlice.Namespace + "/" + epSlice.Name
	l.Sugar().Debugf("update EndpointSlice %s for the service %s", epindex, index)

	s.ebpfServiceLock.Lock()
	defer s.ebpfServiceLock.Unlock()
	if d, ok := s.endpointData[index]; ok {
		if d.Svc != nil {
			if !onlyUpdateTime {
				l.Sugar().Infof("cache the data, and apply new data to ebpf map for the service %v", index)
				oldEps := shallowCopyEdpSliceMap(d.EpsliceList)
				d.EpsliceList[epindex] = epSlice
				if e := s.ebpfhandler.UpdateEbpfMapForService(l, d.Svc, d.Svc, oldEps, d.EpsliceList); e != nil {
					d.EpsliceList = oldEps
					l.Sugar().Errorf("failed to write ebpf map: %v", e)
					return e
				}
			} else {
				l.Sugar().Debugf("just update lastUpdateTime")
				d.EpsliceList[epindex] = epSlice
			}
		} else {
			d.EpsliceList[epindex] = epSlice
			l.Sugar().Debugf("cache the data, but no need to apply new data to ebpf map, cause miss service")
		}
	} else {
		l.Sugar().Debugf("cache the data, but no need to apply new data to ebpf map, cause miss service")
		s.endpointData[index] = &EndpointData{
			Svc: nil,
			EpsliceList: map[string]*discovery.EndpointSlice{
				epindex: epSlice,
			},
		}
	}

	return nil
}

func (s *ebpfWriter) DeleteEndpointSlice(l *zap.Logger, epSlice *discovery.EndpointSlice) error {

	if epSlice == nil {
		return fmt.Errorf("empty service")
	}

	index := k8s.GetEndpointSliceOwnerName(epSlice)
	epindex := epSlice.Namespace + "/" + epSlice.Name
	l.Sugar().Debugf("delete EndpointSlice %s for the service %s", epindex, index)

	s.ebpfServiceLock.Lock()
	defer s.ebpfServiceLock.Unlock()
	if d, ok := s.endpointData[index]; ok {
		if d.Svc == nil {
			// when the service event happens, the data has been removed
			delete(d.EpsliceList, epindex)
		} else {
			if oldEp, ok := d.EpsliceList[epindex]; ok {
				l.Sugar().Infof("delete data from ebpf map for EndpointSlice: %v", index)
				oldEps := shallowCopyEdpSliceMap(d.EpsliceList)
				delete(d.EpsliceList, epindex)
				if e := s.ebpfhandler.UpdateEbpfMapForService(l, d.Svc, d.Svc, oldEps, d.EpsliceList); e != nil {
					d.EpsliceList[epindex] = oldEp
					l.Sugar().Errorf("failed to write ebpf map: %v", e)
					return e
				}
				goto finish
			}
		}
	}
	l.Sugar().Debugf("no need to apply EndpointSlice for ebpf map, cause the data has been already removed")

finish:
	return nil
}

// ---------------------------------------------------------

func (s *ebpfWriter) UpdateNode(l *zap.Logger, node *corev1.Node, onlyUpdateTime bool) error {

	if node == nil {
		return fmt.Errorf("empty node")
	}
	node.ObjectMeta.CreationTimestamp = metav1.Time{
		time.Now(),
	}

	index := node.Name
	l.Sugar().Debugf("update node %s ", index)

	s.ebpfNodeLock.Lock()
	defer s.ebpfNodeLock.Unlock()
	if d, ok := s.nodeData[index]; ok {
		if !onlyUpdateTime {
			l.Sugar().Infof("cache the data, and apply new data to ebpf map for the node %v", index)
			if e := s.ebpfhandler.UpdateEbpfMapForNode(l, d, node); e != nil {
				l.Sugar().Errorf("failed to write ebpf map for the node %v: %v", index, e)
				return e
			}
			s.nodeData[index] = node
		} else {
			l.Sugar().Debugf("just update lastUpdateTime")
			d = node
		}
	} else {
		l.Sugar().Infof("cache the data, and apply new data to ebpf map for the node %v", index)
		if e := s.ebpfhandler.UpdateEbpfMapForNode(l, nil, node); e != nil {
			l.Sugar().Errorf("failed to write ebpf map for the node %v: %v", index, e)
			return e
		}
		s.nodeData[index] = node
	}

	return nil
}

func (s *ebpfWriter) DeleteNode(l *zap.Logger, node *corev1.Node) error {
	if node == nil {
		return fmt.Errorf("empty node")
	}
	index := node.Name
	l.Sugar().Debugf("delete node %s ", index)

	s.ebpfNodeLock.Lock()
	defer s.ebpfNodeLock.Unlock()
	if _, ok := s.nodeData[index]; ok {
		l.Sugar().Infof("delete data from ebpf map for node: %v", index)
		if e := s.ebpfhandler.DeleteEbpfMapForNode(l, node); e != nil {
			l.Sugar().Errorf("failed to write ebpf map for the node %v: %v", index, e)
			return e
		}
		delete(s.nodeData, index)
	} else {
		l.Sugar().Debugf("no need to delete node from ebpf map, cause already removed")
	}
	return nil
}
