// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package ebpfWriter

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/elf-io/balancing/pkg/k8s"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type SvcEndpointData struct {
	Svc *corev1.Service
	// one endpointslice store 100 endpoints by default
	// index: namesapce/endpointSliceName
	EpsliceList map[string]*discovery.EndpointSlice
	// identical to the serviceId in the ebpf map, it is used for event to find policy
	// so only just update ServiceId before updating ebpf map
	ServiceId uint32
}

func shallowCopyEdpSliceMap(t map[string]*discovery.EndpointSlice) map[string]*discovery.EndpointSlice {
	m := make(map[string]*discovery.EndpointSlice)
	for k, v := range t {
		m[k] = v
	}
	return m
}

func (s *ebpfWriter) UpdateServiceByService(l *zap.Logger, svc *corev1.Service, onlyUpdateTime bool) error {

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
	if d, ok := s.serviceData[index]; ok {
		if d.EpsliceList != nil && len(d.EpsliceList) > 0 {
			if !onlyUpdateTime {
				l.Sugar().Infof("cache the data, and apply new data to ebpf map for service %v", index)
				// only before update ebpf map, update serviceId
				t, ipv6flag := ebpf.GenerateSvcV4Id(svc)
				if t == 0 {
					if ipv6flag {
						l.Sugar().Debug("ignore ipv6 service ")
					} else {
						l.Sugar().Errorf("failed to get serviceV4Id for service :%+v ", svc)
					}
					return fmt.Errorf("failed to get serviceV4Id for service  ")
				}
				if d.ServiceId != 0 && d.ServiceId != t {
					l.Sugar().Warnf("the serviceId of service %s/%s changes from %d to %d  ", svc.Namespace, svc.Name, d.ServiceId, t)
				} else if d.ServiceId == 0 {
					l.Sugar().Infof("update ServiceId to %d", t)
				}
				d.ServiceId = t
				//
				if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_SERVICE, d.Svc, svc, d.EpsliceList, d.EpsliceList, nil); e != nil {
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
		t := &SvcEndpointData{
			Svc:         svc,
			EpsliceList: make(map[string]*discovery.EndpointSlice),
		}
		s.serviceData[index] = t
	}

	return nil
}

func (s *ebpfWriter) DeleteServiceByService(l *zap.Logger, svc *corev1.Service) error {
	if svc == nil {
		return fmt.Errorf("empty service")
	}

	index := svc.Namespace + "/" + svc.Name
	l.Sugar().Debugf("delete service %s", index)

	s.ebpfServiceLock.Lock()
	defer s.ebpfServiceLock.Unlock()
	if d, ok := s.serviceData[index]; ok {
		// todo : generate a ebpf map data and apply it
		l.Sugar().Infof("delete data from ebpf map for service: %v", index)
		if e := s.ebpfhandler.DeleteEbpfMapForService(l, ebpf.NAT_TYPE_SERVICE, d.Svc, d.EpsliceList, nil); e != nil {
			l.Sugar().Errorf("failed to write ebpf map: %v", e)
			return e
		}
		delete(s.serviceData, index)
	} else {
		l.Sugar().Debugf("no need to delete service from ebpf map, cause already removed")
	}

	return nil
}

// -------------------------------------------------------------
func (s *ebpfWriter) UpdateServiceByEndpointSlice(l *zap.Logger, epSlice *discovery.EndpointSlice, onlyUpdateTime bool) error {

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
	if d, ok := s.serviceData[index]; ok {
		if d.Svc != nil {
			if !onlyUpdateTime {
				l.Sugar().Infof("cache the data, and apply new data to ebpf map for the service %v", index)
				oldEps := shallowCopyEdpSliceMap(d.EpsliceList)
				d.EpsliceList[epindex] = epSlice
				// it only updates serviceId before updating ebpf map
				t, ipv6flag := ebpf.GenerateSvcV4Id(d.Svc)
				if t == 0 {
					if ipv6flag {
						l.Sugar().Debug("ignore ipv6 service ")
					} else {
						l.Sugar().Errorf("failed to get serviceV4Id for service :%+v ", d.Svc)
					}
					return fmt.Errorf("failed to get serviceV4Id for service  ")
				}
				if d.ServiceId != 0 && d.ServiceId != t {
					l.Sugar().Warnf("the serviceId of service %s/%s changes from %d to %d  ", d.Svc.Namespace, d.Svc.Name, d.ServiceId, t)
				} else if d.ServiceId == 0 {
					l.Sugar().Infof("update ServiceId to %d", t)
				}
				d.ServiceId = t
				// update
				if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_SERVICE, d.Svc, d.Svc, oldEps, d.EpsliceList, nil); e != nil {
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
		s.serviceData[index] = &SvcEndpointData{
			Svc: nil,
			EpsliceList: map[string]*discovery.EndpointSlice{
				epindex: epSlice,
			},
		}
	}

	return nil
}

func (s *ebpfWriter) DeleteServiceByEndpointSlice(l *zap.Logger, epSlice *discovery.EndpointSlice) error {

	if epSlice == nil {
		return fmt.Errorf("empty service")
	}

	index := k8s.GetEndpointSliceOwnerName(epSlice)
	epindex := epSlice.Namespace + "/" + epSlice.Name
	l.Sugar().Debugf("delete EndpointSlice %s for the service %s", epindex, index)

	s.ebpfServiceLock.Lock()
	defer s.ebpfServiceLock.Unlock()
	if d, ok := s.serviceData[index]; ok {
		if d.Svc == nil {
			// when the service event happens, the data has been removed
			delete(d.EpsliceList, epindex)
		} else {
			if oldEp, ok := d.EpsliceList[epindex]; ok {
				l.Sugar().Infof("delete data from ebpf map for EndpointSlice: %v", index)
				oldEps := shallowCopyEdpSliceMap(d.EpsliceList)
				delete(d.EpsliceList, epindex)
				// only before update ebpf map, update serviceId
				t, ipv6flag := ebpf.GenerateSvcV4Id(d.Svc)
				if t == 0 {
					if ipv6flag {
						l.Sugar().Debug("ignore ipv6 service ")
					} else {
						l.Sugar().Errorf("failed to get serviceV4Id for service :%+v ", d.Svc)
					}
					return fmt.Errorf("failed to get serviceV4Id for service  ")
				}
				if d.ServiceId != 0 && d.ServiceId != t {
					l.Sugar().Warnf("the serviceId of service %s/%s changes from %d to %d  ", d.Svc.Namespace, d.Svc.Name, d.ServiceId, t)
				} else if d.ServiceId == 0 {
					l.Sugar().Infof("update ServiceId to %d", t)
				}
				d.ServiceId = t
				// update
				if e := s.ebpfhandler.UpdateEbpfMapForService(l, ebpf.NAT_TYPE_SERVICE, d.Svc, d.Svc, oldEps, d.EpsliceList, nil); e != nil {
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
