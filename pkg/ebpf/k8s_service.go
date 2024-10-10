package ebpf

import (
	"encoding/binary"
	"fmt"
	"github.com/elf-io/balancing/pkg/nodeId"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"net"
	"reflect"
)

// map the key and value of the service map
type serviceMapData struct {
	key *bpf_cgroupMapkeyService
	val *bpf_cgroupMapvalueService
}

// map the key and value of the backend map
type backendMapData struct {
	key *bpf_cgroupMapkeyBackend
	val *bpf_cgroupMapvalueBackend
}

func buildEbpfMapDataForV4Service(natType uint8, svc *corev1.Service, edsList map[string]*discovery.EndpointSlice) ([]*serviceMapData, []*backendMapData, error) {

	if svc == nil {
		return nil, nil, fmt.Errorf("service is empty")
	}

	resultSvcList := []*serviceMapData{}
	resultBackList := []*backendMapData{}

	svcV4Id := GenerateSvcV4Id(svc)
	if svcV4Id == 0 {
		return nil, nil, fmt.Errorf("failed to generate svcId")
	}
	affinityTime := GetServiceAffinityTime(svc)
	serviceFlags := GetServiceFlag(svc)

	for _, svcPort := range svc.Spec.Ports {

		protocol := GetPortProtocol(&svcPort)

		// -------------------  generate data for backend map --------------------
		// 1组 service port， 对应 一组 clusterIpPort backend + nodePort backend
		localEp, remoteEp := ClassifyV4Endpoint(edsList)
		allEp := []*discovery.Endpoint{}
		allEp = append(allEp, localEp...)
		allEp = append(allEp, remoteEp...)
		for order, edp := range allEp {
			// for pod port
			backMapKey := bpf_cgroupMapkeyBackend{
				Order:   uint32(order),
				SvcId:   svcV4Id,
				Dport:   uint16(svcPort.Port),
				Proto:   protocol,
				NatType: natType,
				Scope:   SCOPE_LOCAL_CLUSTER,
			}
			nodeid := uint32(0)
			if edp.NodeName != nil && len(*edp.NodeName) > 0 {
				nodeid, _ = nodeId.NodeIdManagerHander.GetNodeId(*(edp.NodeName))
			}
			backMapVal := bpf_cgroupMapvalueBackend{
				PodAddress: GetEndpointIPv4Address(edp),
				NodeId:     nodeid,
				PodPort:    uint16(svcPort.TargetPort.IntValue()),
				NodePort:   uint16(svcPort.NodePort),
			}
			resultBackList = append(resultBackList, &backendMapData{
				key: &backMapKey,
				val: &backMapVal,
			})

			// for nodePort
			if svcPort.NodePort != 0 && natType == NAT_TYPE_SERVICE {
				m := backMapKey
				m.Dport = uint16(svcPort.NodePort)
				n := backMapVal
				resultBackList = append(resultBackList, &backendMapData{
					key: &m,
					val: &n,
				})
			}
		}

		// -------------------  generate data for service map --------------------
		addFunc := func(vip net.IP, port uint16, natType, natMode uint8) {
			svcMapKey := bpf_cgroupMapkeyService{
				Address: binary.LittleEndian.Uint32(vip.To4()),
				Dport:   port,
				Proto:   protocol,
				NatType: natType,
				Scope:   SCOPE_LOCAL_CLUSTER,
			}
			svcMapVal := bpf_cgroupMapvalueService{
				SvcId:             svcV4Id,
				TotalBackendCount: uint32(len(allEp)),
				LocalBackendCount: uint32(len(localEp)),
				AffinitySecond:    affinityTime,
				ServiceFlags:      serviceFlags,
				BalancingFlags:    0,
				RedirectFlags:     0,
				NatMode:           natMode,
			}
			resultSvcList = append(resultSvcList, &serviceMapData{
				key: &svcMapKey,
				val: &svcMapVal,
			})
		}

		// add date for clusterip
		for _, vip := range getClusterIPs(svc, corev1.IPv4Protocol) {
			addFunc(vip, uint16(svcPort.Port), natType, NatModeServiceClusterip)
		}

		if natType == NAT_TYPE_SERVICE {
			// add date for loadbalancerIP
			for _, vip := range GetServiceV4LoadbalancerIP(svc) {
				addFunc(vip, uint16(svcPort.Port), natType, NatModeServiceLoadBalancer)
			}

			// add date for ExternalIPs
			if svc.Spec.ExternalIPs != nil {
				for _, v := range svc.Spec.ExternalIPs {
					vip := net.ParseIP(v)
					if vip.To4() != nil {
						addFunc(vip.To4(), uint16(svcPort.Port), natType, NatModeServiceExternalIp)
					}
				}
			}

			// add date for NodePort
			// handle nodePort alone cause it uses nodePort
			if svcPort.NodePort != 0 {
				// generate data for service map
				addFunc(NODEPORT_V4_IP, uint16(svcPort.NodePort), natType, NatModeServiceNodePort)
			}
		}
	}

	return resultSvcList, resultBackList, nil
}

// -----------------------------------------------

func (s *EbpfProgramStruct) applyEpfMapDataV4Service(l *zap.Logger, oldList, newList []*serviceMapData) error {

	delKeyList := []bpf_cgroupMapkeyService{}
	addKeyList := []bpf_cgroupMapkeyService{}
	addValList := []bpf_cgroupMapvalueService{}

	l.Sugar().Debugf("service map %d items in oldList: ", len(oldList))
	for k, v := range oldList {
		l.Sugar().Debugf("service map oldList[%d]: key=%s, value=%s ", k, *v.key, *v.val)
	}
	l.Sugar().Debugf("service map %d items in newList: ", len(newList))
	for k, v := range newList {
		l.Sugar().Debugf("service map newList[%d]: key=%s, value=%s ", k, *v.key, *v.val)
	}

OUTER_OLD:
	for _, oldKey := range oldList {
		for _, newKey := range newList {
			if reflect.DeepEqual(oldKey.key, newKey.key) {
				if !reflect.DeepEqual(oldKey.val, newKey.val) {
					addKeyList = append(addKeyList, *newKey.key)
					addValList = append(addValList, *newKey.val)
					l.Sugar().Infof("ebpf map of the service updates: key=%s , value=%s ", newKey.key, newKey.val)
				}
				continue OUTER_OLD
			}
		}
		l.Sugar().Infof("ebpf map of the service deletes: key=%s , value=%s ", oldKey.key, oldKey.val)
		delKeyList = append(delKeyList, *oldKey.key)
	}

OUTER_NEW:
	for _, newKey := range newList {
		for _, oldKey := range oldList {
			if reflect.DeepEqual(oldKey.key, newKey.key) {
				continue OUTER_NEW
			}
		}
		addKeyList = append(addKeyList, *newKey.key)
		addValList = append(addValList, *newKey.val)
		l.Sugar().Infof("ebpf map of the service updates: key=%s , value=%s ", newKey.key, newKey.val)
	}

	// -------- apply
	// must deletion first, then apply new .
	if len(delKeyList) > 0 {
		if e := s.DeleteMapService(delKeyList); e != nil {
			l.Sugar().Errorf("failed to delete service map: %v", e)
			return fmt.Errorf("failed to delete service map: %v", e)
		}
		l.Sugar().Infof("succeeded to delete %d items in service data ", len(delKeyList))
	}
	if len(addKeyList) > 0 {
		if e := s.UpdateMapService(addKeyList, addValList); e != nil {
			l.Sugar().Errorf("failed to update service map: %v", e)
			return fmt.Errorf("failed to update service map: %v", e)
		}
		l.Sugar().Infof("succeeded to update %d items in service map: ", len(addKeyList))
	}

	return nil
}

func (s *EbpfProgramStruct) applyEpfMapDataV4Backend(l *zap.Logger, oldList, newList []*backendMapData) error {

	delKeyList := []bpf_cgroupMapkeyBackend{}
	addKeyList := []bpf_cgroupMapkeyBackend{}
	addValList := []bpf_cgroupMapvalueBackend{}

	l.Sugar().Debugf("backend map %d items in oldList: ", len(oldList))
	for k, v := range oldList {
		l.Sugar().Debugf("backend map oldList[%d]: key=%s, value=%s ", k, *v.key, *v.val)
	}
	l.Sugar().Debugf("backend map %d items in newList: ", len(newList))
	for k, v := range newList {
		l.Sugar().Debugf("backend map newList[%d]: key=%s, value=%s ", k, *v.key, *v.val)
	}

OUTER_OLD:
	for _, oldKey := range oldList {
		for _, newKey := range newList {
			if reflect.DeepEqual(oldKey.key, newKey.key) {
				if !reflect.DeepEqual(oldKey.val, newKey.val) {
					addKeyList = append(addKeyList, *newKey.key)
					addValList = append(addValList, *newKey.val)
					l.Sugar().Infof("ebpf map of the backend updates: key=%s , value=%s ", newKey.key, newKey.val)
				}
				continue OUTER_OLD
			}
		}
		l.Sugar().Infof("ebpf map of the backend deletes: key=%s , value=%s ", oldKey.key, oldKey.val)
		delKeyList = append(delKeyList, *oldKey.key)
	}

OUTER_NEW:
	for _, newKey := range newList {
		for _, oldKey := range oldList {
			if reflect.DeepEqual(oldKey.key, newKey.key) {
				continue OUTER_NEW
			}
		}
		addKeyList = append(addKeyList, *newKey.key)
		addValList = append(addValList, *newKey.val)
		l.Sugar().Infof("ebpf map of the backend updates: key=%s , value=%s ", newKey.key, newKey.val)
	}

	// -------- apply
	// must deletion first, then apply new
	if len(delKeyList) > 0 {
		if e := s.DeleteMapBackend(delKeyList); e != nil {
			l.Sugar().Errorf("failed to delete backend map: %v", e)
			return fmt.Errorf("failed to delete backend map: %v", e)
		}
		l.Sugar().Infof("succeeded to delete %d items in backend data ", len(delKeyList))
	}

	if len(addKeyList) > 0 {
		if e := s.UpdateMapBackend(addKeyList, addValList); e != nil {
			l.Sugar().Errorf("failed to update backend map: %v", e)
			return fmt.Errorf("failed to update backend map: %v", e)
		}
		l.Sugar().Infof("succeeded to update %d items in backend map: ", len(addKeyList))
	}

	return nil
}

// -------------------------------------------------- for k8s service , localRedirect, balancing

func (s *EbpfProgramStruct) UpdateEbpfMapForService(l *zap.Logger, natType uint8, oldSvc, newSvc *corev1.Service, oldEdsList, newEdsList map[string]*discovery.EndpointSlice) error {

	processIpv4 := false
	processIpv6 := false
	for _, v := range newSvc.Spec.IPFamilies {
		if v == corev1.IPv4Protocol {
			processIpv4 = true
		}
		if v == corev1.IPv6Protocol {
			processIpv6 = true
		}
	}

	if processIpv4 {
		oldSvcList := []*serviceMapData{}
		oldBkList := []*backendMapData{}
		var err1 error
		if oldSvc != nil && len(oldEdsList) > 0 {
			oldSvcList, oldBkList, err1 = buildEbpfMapDataForV4Service(natType, oldSvc, oldEdsList)
			if err1 != nil {
				return fmt.Errorf("failed to buildEbpfMapDataForV4Service: %v", err1)
			}
		}

		newSvcList, newBkList, err2 := buildEbpfMapDataForV4Service(natType, newSvc, newEdsList)
		if err2 != nil {
			return fmt.Errorf("failed to buildEbpfMapDataForV4Service: %v", err2)
		}

		if e := s.applyEpfMapDataV4Service(l, oldSvcList, newSvcList); e != nil {
			return fmt.Errorf("failed to applyEpfMapDataV4Service: %v", e)
		}
		if e := s.applyEpfMapDataV4Backend(l, oldBkList, newBkList); e != nil {
			return fmt.Errorf("failed to applyEpfMapDataV4Backend: %v", e)
		}
	}

	if processIpv6 {
		l.Sugar().Infof("does not suppport ipv6, abandon applying ")
	}

	return nil
}

func (s *EbpfProgramStruct) DeleteEbpfMapForService(l *zap.Logger, natType uint8, svc *corev1.Service, edsList map[string]*discovery.EndpointSlice) error {

	processIpv4 := false
	processIpv6 := false
	for _, v := range svc.Spec.IPFamilies {
		if v == corev1.IPv4Protocol {
			processIpv4 = true
		}
		if v == corev1.IPv6Protocol {
			processIpv6 = true
		}
	}

	if processIpv4 {
		svcList, bkList, err := buildEbpfMapDataForV4Service(natType, svc, edsList)
		if err != nil {
			return fmt.Errorf("failed to buildEbpfMapDataForV4Service: %v", err)
		}

		if e := s.applyEpfMapDataV4Service(l, svcList, []*serviceMapData{}); e != nil {
			return fmt.Errorf("failed to applyEpfMapDataV4Service: %v", e)
		}
		if e := s.applyEpfMapDataV4Backend(l, bkList, []*backendMapData{}); e != nil {
			return fmt.Errorf("failed to applyEpfMapDataV4Backend: %v", e)
		}
	}

	if processIpv6 {
		l.Sugar().Infof("does not suppport ipv6, abandon applying ")
	}

	return nil
}
