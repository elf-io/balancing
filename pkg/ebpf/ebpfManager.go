// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package ebpf

// refer to https://github.com/cilium/pwru/blob/main/build.go
//go:generate sh -c "echo Generating for $TARGETARCH"
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target $TARGETARCH -cc clang -no-strip bpf_cgroup  bpf/cgroup.c

// .. //go:generate go run github.com/cilium/ebpf/cmd/bpf2go -no-strip -cc clang -target bpf -cflags "-D__TARGET_ARCH_x86"  bpf_cgroup   bpf/cgroup.c

import (
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/elf-io/balancing/pkg/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"path/filepath"
	"strings"

	// "github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"os"
)

// -------------------------------------------
const (
	EventChanLength = 1000
)

// -----------------------------------

type EbpfMaps struct {
	MapAffinity    *ebpf.Map
	MapBackend     *ebpf.Map
	MapEvent       *ebpf.Map
	MapNatRecord   *ebpf.Map
	MapNodeIp      *ebpf.Map
	MapNodeProxyIp *ebpf.Map
	MapService     *ebpf.Map
	MapConfigure   *ebpf.Map
}

type EbpfProgramStruct struct {
	BpfObjCgroup      bpf_cgroupObjects
	CgroupLinkConnect link.Link
	CgroupLinkSend    link.Link
	CgroupLinkRecv    link.Link
	CgroupLinkPeer    link.Link
	Event             chan MapEventValue
	l                 *zap.Logger

	// for debug cli to load map alone
	EbpfMaps *EbpfMaps
}

type EbpfProgram interface {
	// load the ebpf program and map
	LoadProgramp() error
	UnloadProgramp() error

	// for debug cli to load pinned map
	LoadAllEbpfMap(string) error
	UnloadAllEbpfMap()

	GetMapDataEvent() <-chan MapEventValue

	// for debug cli
	PrintMapService(*uint8, *uint32) error
	PrintMapNodeIp() error
	PrintMapNodeProxyIp() error
	PrintMapBackend(*uint8, *uint32) error
	PrintMapAffinity() error
	PrintMapNatRecord() error
	PrintMapConfigure() error

	CleanMapService() (int, error)
	CleanMapNodeIp() (int, error)
	CleanMapNodeProxyIp() (int, error)
	CleanMapBackend() (int, error)
	CleanMapAffinity() (int, error)
	CleanMapNatRecord() (int, error)

	UpdateMapService([]bpf_cgroupMapkeyService, []bpf_cgroupMapvalueService) error
	UpdateMapBackend([]bpf_cgroupMapkeyBackend, []bpf_cgroupMapvalueBackend) error
	UpdateMapNodeIp([]bpf_cgroupMapkeyNodeIp, []uint32) error
	UpdateMapNodeProxyIp([]uint32, []bpf_cgroupMapvalueNodeProxyIp) error
	UpdateMapAffinity([]bpf_cgroupMapkeyAffinity, []bpf_cgroupMapvalueAffinity) error
	UpdateMapNatRecord([]bpf_cgroupMapkeyNatRecord, []bpf_cgroupMapvalueNatRecord) error
	UpdateMapConfigure(uint32, uint32) error

	DeleteMapNatRecord([]bpf_cgroupMapkeyNatRecord) error
	DeleteMapAffinity([]bpf_cgroupMapkeyAffinity) error
	DeleteMapNodeIp([]bpf_cgroupMapkeyNodeIp) error
	DeleteMapNodeProxyIp([]uint32) error
	DeleteMapService([]bpf_cgroupMapkeyService) error
	DeleteMapBackend([]bpf_cgroupMapkeyBackend) error

	// for agent
	// for k8s service and endpointslice
	UpdateEbpfMapForService(*zap.Logger, uint8, *corev1.Service, *corev1.Service, map[string]*discovery.EndpointSlice, map[string]*discovery.EndpointSlice, *uint8) error
	DeleteEbpfMapForService(*zap.Logger, uint8, *corev1.Service, map[string]*discovery.EndpointSlice, *uint8) error
	// for k8s node
	UpdateEbpfMapForNode(*zap.Logger, *corev1.Node, *corev1.Node) error
	DeleteEbpfMapForNode(*zap.Logger, *corev1.Node) error
}

var _ EbpfProgram = &EbpfProgramStruct{}

// ------------------------------------

func NewEbpfProgramMananger(l *zap.Logger) EbpfProgram {
	return &EbpfProgramStruct{
		l: l,
	}
}

func (s *EbpfProgramStruct) LoadProgramp() error {

	s.Event = make(chan MapEventValue, EventChanLength)

	if err := checkOrMountBpfFs(types.BpfFSPath); err != nil {
		return fmt.Errorf("failed to mount bpf fs: %v", err)
	}

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("failed to RemoveMemlock: %v", err)
	}

	// attach to cgroup
	// sync.Once.Do(func() {
	if err := checkOrMountCgroupV2(types.CgroupV2Path); err != nil {
		return fmt.Errorf("failed to mount cgroup v2: %s", err)
	}
	// })

	// create the directory for map pin path
	if stat, err := os.Stat(types.MapsPinpath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(types.MapsPinpath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to mkdir %s: %v", types.MapsPinpath, err)
			}
		} else {
			return fmt.Errorf("Failed to stat the path %s: %w", types.MapsPinpath, err)
		}
	} else {
		if !stat.IsDir() {
			return fmt.Errorf("%s is a file which is not a directory", types.MapsPinpath)
		}
	}

	// 这个函数( loadBpf_xxxObjects )是 bpf2go 生成的 go 文件中 加载 ebpf 程序到内核
	err := loadBpf_cgroupObjects(&s.BpfObjCgroup, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: types.MapsPinpath,
		}})
	if err != nil {
		return fmt.Errorf("failed to load ebpf obj: %v", err)
	}

	// 把 ebpf 程序再挂载到 cgroup
	// https://github.com/cilium/ebpf/blob/main/link/cgroup.go#L43
	s.l.Sugar().Debugf("attach AttachCGroupUDP4Sendmsg")
	// 返回的 CgroupLink 必须进行存储，否则被 GC 而触发 attach 被释放的问题
	s.CgroupLinkSend, err = link.AttachCgroup(link.CgroupOptions{
		Path:    types.CgroupV2Path,
		Attach:  ebpf.AttachCGroupUDP4Sendmsg,
		Program: s.BpfObjCgroup.bpf_cgroupPrograms.Sock4Sendmsg,
	})
	if err != nil {
		return fmt.Errorf("Error attaching Sock4Sendmsg to cgroup: %v", err)
	}
	s.l.Sugar().Debugf("attach AttachCGroupUDP4Recvmsg")
	// 返回的 CgroupLink 必须进行存储，否则会被 GC 而触发 attach 被释放的问题
	s.CgroupLinkRecv, err = link.AttachCgroup(link.CgroupOptions{
		Path:    types.CgroupV2Path,
		Attach:  ebpf.AttachCGroupUDP4Recvmsg,
		Program: s.BpfObjCgroup.bpf_cgroupPrograms.Sock4Recvmsg,
	})
	if err != nil {
		return fmt.Errorf("Error attaching Sock4Recvmsg to cgroup: %v", err)
	}
	s.l.Sugar().Debugf("attach AttachCgroupInet4GetPeername")
	// 返回的 CgroupLink 必须进行存储，否则会被 GC 而触发 attach 被释放的问题
	s.CgroupLinkPeer, err = link.AttachCgroup(link.CgroupOptions{
		Path:    types.CgroupV2Path,
		Attach:  ebpf.AttachCgroupInet4GetPeername,
		Program: s.BpfObjCgroup.bpf_cgroupPrograms.Sock4Getpeername,
	})
	if err != nil {
		return fmt.Errorf("Error attaching Sock4Getpeername to cgroup: %v", err)
	}
	s.l.Sugar().Debugf("attach AttachCGroupInet4Connect")
	// 返回的 CgroupLink 必须进行存储，否则会被 GC 而触发 attach 被释放的问题
	s.CgroupLinkConnect, err = link.AttachCgroup(link.CgroupOptions{
		Path:    types.CgroupV2Path,
		Attach:  ebpf.AttachCGroupInet4Connect,
		Program: s.BpfObjCgroup.bpf_cgroupPrograms.Sock4Connect,
	})
	if err != nil {
		return fmt.Errorf("Error attaching Sock4Connect to cgroup: %v", err)
	}

	go s.daemonGetEvent()

	// set configuration
	if strings.ToLower(types.AgentConfig.EbpfLogLevel) == types.LogLevelEbpfDebug {
		s.l.Sugar().Infof("ebpf debug level: verbose")
		if err := s.UpdateMapConfigure(MapConfigureKeyIndexDebugLevel, MapConfigureValueDebugLevelVerbose); err != nil {
			s.l.Sugar().Fatalf("%v", err)
		}
	} else if strings.ToLower(types.AgentConfig.EbpfLogLevel) == types.LogLevelEbpfInfo {
		s.l.Sugar().Infof("ebpf debug level: info")
		if err := s.UpdateMapConfigure(MapConfigureKeyIndexDebugLevel, MapConfigureValueDebugLevelInfo); err != nil {
			s.l.Sugar().Fatalf("%v", err)
		}
	} else {
		s.l.Sugar().Infof("ebpf debug level: error")
		if err := s.UpdateMapConfigure(MapConfigureKeyIndexDebugLevel, MapConfigureValueDebugLevelError); err != nil {
			s.l.Sugar().Fatalf("%v", err)
		}
	}
	if types.AgentConfig.Configmap.EnableIPv4 {
		s.l.Sugar().Infof("ebpf ipv4 enabled: true")
		if err := s.UpdateMapConfigure(MapConfigureKeyIndexIpv4Enabled, MapConfigureValueEnabled); err != nil {
			s.l.Sugar().Fatalf("%v", err)
		}
	} else {
		s.l.Sugar().Infof("ebpf ipv4 enabled: false")
		if err := s.UpdateMapConfigure(MapConfigureKeyIndexIpv4Enabled, MapConfigureValueDisabled); err != nil {
			s.l.Sugar().Fatalf("%v", err)
		}
	}
	if types.AgentConfig.Configmap.EnableIPv6 {
		s.l.Sugar().Infof("ebpf ipv6 enabled: true")
		if err := s.UpdateMapConfigure(MapConfigureKeyIndexIpv6Enabled, MapConfigureValueEnabled); err != nil {
			s.l.Sugar().Fatalf("%v", err)
		}
	} else {
		s.l.Sugar().Infof("ebpf ipv6 enabled: false")
		if err := s.UpdateMapConfigure(MapConfigureKeyIndexIpv6Enabled, MapConfigureValueDisabled); err != nil {
			s.l.Sugar().Fatalf("%v", err)
		}
	}

	return nil
}

func (s *EbpfProgramStruct) UnloadProgramp() error {

	if s.CgroupLinkConnect != nil {
		fmt.Printf("Closing  cgroup v2 ...\n")
		s.CgroupLinkConnect.Close()
	}
	if s.CgroupLinkSend != nil {
		fmt.Printf("Closing  cgroup v2 ...\n")
		s.CgroupLinkSend.Close()
	}
	if s.CgroupLinkRecv != nil {
		fmt.Printf("Closing  cgroup v2 ...\n")
		s.CgroupLinkRecv.Close()
	}
	if s.CgroupLinkPeer != nil {
		fmt.Printf("Closing  cgroup v2 ...\n")
		s.CgroupLinkPeer.Close()
	}

	// unping and close ebpf map
	if s.BpfObjCgroup.bpf_cgroupMaps.MapBackend != nil {
		if err := s.BpfObjCgroup.bpf_cgroupMaps.MapBackend.Unpin(); err != nil {
			// 处理错误
			fmt.Println("Error:", err)
		}
	}
	if s.BpfObjCgroup.bpf_cgroupMaps.MapService != nil {
		if err := s.BpfObjCgroup.bpf_cgroupMaps.MapService.Unpin(); err != nil {
			// 处理错误
			fmt.Println("Error:", err)
		}
	}
	if s.BpfObjCgroup.bpf_cgroupMaps.MapAffinity != nil {
		if err := s.BpfObjCgroup.bpf_cgroupMaps.MapAffinity.Unpin(); err != nil {
			// 处理错误
			fmt.Println("Error:", err)
		}
	}
	if s.BpfObjCgroup.bpf_cgroupMaps.MapNodeIp != nil {
		if err := s.BpfObjCgroup.bpf_cgroupMaps.MapNodeIp.Unpin(); err != nil {
			// 处理错误
			fmt.Println("Error:", err)
		}
	}
	if s.BpfObjCgroup.bpf_cgroupMaps.MapNodeProxyIp != nil {
		if err := s.BpfObjCgroup.bpf_cgroupMaps.MapNodeProxyIp.Unpin(); err != nil {
			// 处理错误
			fmt.Println("Error:", err)
		}
	}
	if s.BpfObjCgroup.bpf_cgroupMaps.MapNatRecord != nil {
		if err := s.BpfObjCgroup.bpf_cgroupMaps.MapNatRecord.Unpin(); err != nil {
			// 处理错误
			fmt.Println("Error:", err)
		}
	}
	if s.BpfObjCgroup.bpf_cgroupMaps.MapEvent != nil {
		if err := s.BpfObjCgroup.bpf_cgroupMaps.MapEvent.Unpin(); err != nil {
			// 处理错误
			fmt.Println("Error:", err)
		}
	}

	fmt.Printf("Closing progs ...\n")
	s.BpfObjCgroup.bpf_cgroupPrograms.Close()
	s.BpfObjCgroup.bpf_cgroupMaps.Close()

	s.BpfObjCgroup.Close()

	return nil
}

func (s *EbpfProgramStruct) LoadAllEbpfMap(mapPinDir string) error {

	if s.EbpfMaps != nil {
		// already load
		return nil
	}

	s.EbpfMaps = &EbpfMaps{}

	var err error
	mapdir := mapPinDir
	if len(mapdir) == 0 {
		mapdir = types.MapsPinpath
	}

	f := filepath.Join(mapdir, "map_affinity")
	s.EbpfMaps.MapAffinity, err = ebpf.LoadPinnedMap(f, &ebpf.LoadPinOptions{})
	if err != nil {
		s.UnloadAllEbpfMap()
		return fmt.Errorf("failed to load map %s\n", f)
	}

	f = filepath.Join(mapdir, "map_configure")
	s.EbpfMaps.MapConfigure, err = ebpf.LoadPinnedMap(f, &ebpf.LoadPinOptions{})
	if err != nil {
		s.UnloadAllEbpfMap()
		return fmt.Errorf("failed to load map %s\n", f)
	}

	f = filepath.Join(mapdir, "map_backend")
	s.EbpfMaps.MapBackend, err = ebpf.LoadPinnedMap(f, &ebpf.LoadPinOptions{})
	if err != nil {
		s.UnloadAllEbpfMap()
		return fmt.Errorf("failed to load map %s\n", f)
	}

	f = filepath.Join(mapdir, "map_event")
	s.EbpfMaps.MapEvent, err = ebpf.LoadPinnedMap(f, &ebpf.LoadPinOptions{})
	if err != nil {
		s.UnloadAllEbpfMap()
		return fmt.Errorf("failed to load map %s\n", f)
	}

	f = filepath.Join(mapdir, "map_nat_record")
	s.EbpfMaps.MapNatRecord, err = ebpf.LoadPinnedMap(f, &ebpf.LoadPinOptions{})
	if err != nil {
		s.UnloadAllEbpfMap()
		return fmt.Errorf("failed to load map %s\n", f)
	}

	f = filepath.Join(mapdir, "map_node_ip")
	s.EbpfMaps.MapNodeIp, err = ebpf.LoadPinnedMap(f, &ebpf.LoadPinOptions{})
	if err != nil {
		s.UnloadAllEbpfMap()
		return fmt.Errorf("failed to load map %s\n", f)
	}

	f = filepath.Join(mapdir, "map_node_proxy_ip")
	s.EbpfMaps.MapNodeProxyIp, err = ebpf.LoadPinnedMap(f, &ebpf.LoadPinOptions{})
	if err != nil {
		s.UnloadAllEbpfMap()
		return fmt.Errorf("failed to load map %s\n", f)
	}

	f = filepath.Join(mapdir, "map_service")
	s.EbpfMaps.MapService, err = ebpf.LoadPinnedMap(f, &ebpf.LoadPinOptions{})
	if err != nil {
		s.UnloadAllEbpfMap()
		return fmt.Errorf("failed to load map %s\n", f)
	}

	return nil
}

func (s *EbpfProgramStruct) UnloadAllEbpfMap() {
	if s.EbpfMaps == nil {
		// already load
		return
	}
	if s.EbpfMaps.MapAffinity != nil {
		s.EbpfMaps.MapAffinity.Close()
	}
	if s.EbpfMaps.MapBackend != nil {
		s.EbpfMaps.MapBackend.Close()
	}
	if s.EbpfMaps.MapEvent != nil {
		s.EbpfMaps.MapEvent.Close()
	}
	if s.EbpfMaps.MapNatRecord != nil {
		s.EbpfMaps.MapNatRecord.Close()
	}
	if s.EbpfMaps.MapNodeIp != nil {
		s.EbpfMaps.MapNodeIp.Close()
	}
	if s.EbpfMaps.MapNodeProxyIp != nil {
		s.EbpfMaps.MapNodeProxyIp.Close()
	}
	if s.EbpfMaps.MapService != nil {
		s.EbpfMaps.MapService.Close()
	}
	s.EbpfMaps = nil
}

// ------------------------------------------- map

// get data from map
// func (s *EbpfProgramStruct) daemonGetEvent() {
//
// 	rd, err := ringbuf.NewReader(s.BpfObjCgroup.bpf_cgroupMaps.MapEvent)
// 	if err != nil {
// 		fmt.Printf("failed to read ebpf map : %v ", err)
// 		os.Exit(1)
// 	}
// 	defer rd.Close()
//
// 	for {
// 		record, err := rd.Read()
// 		if err != nil {
// 			if errors.Is(err, ringbuf.ErrClosed) {
// 				fmt.Printf("received signal, exiting reading events..\n")
// 				break
// 			}
// 			fmt.Printf("failed to read event: %v", err)
// 			continue
// 		}
//
// 		t := MapEventValue{}
// 		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.NativeEndian, &t); err != nil {
// 			fmt.Printf("parsing ringbuf event: %s", err)
// 			continue
// 		}
// 		// fmt.Printf("get event data: %v \n", t)
//
// 		select {
// 		case s.Event <- t:
// 		default:
// 			fmt.Printf("error, failed to write data to event chan, miss data: %v \n", t)
// 		}
// 	}
//
// }

// -----------------------

// func (s *EbpfProgramStruct) DeleteMapDataFloatip(keyList []bpf_cgroupMapkeyFloatipV4) error {
// 	if s.BpfObjCgroup.MapFloatipV4 == nil {
// 		return fmt.Errorf("ebpf map does not exist")
// 	}
// 	if keyList != nil && len(keyList) == 0 {
// 		return fmt.Errorf("empty key")
// 	}
//
// 	if keyList == nil {
// 		// delete all
// 		keys := make([]bpf_cgroupMapkeyFloatipV4, 100)
// 		vals := make([]bpf_cgroupMapvlaueFloatipV4, 100)
//
// 		var cursor ebpf.MapBatchCursor
// 		count := 0
// 		for {
// 			c, batchErr := s.BpfObjCgroup.MapFloatipV4.BatchLookup(&cursor, keys, vals, nil)
// 			count += c
// 			if batchErr != nil {
// 				if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
// 					// end
// 					break
// 				}
// 				return fmt.Errorf("failed to batchlookup , reason: %v ", batchErr)
// 			}
// 			if _, batchErr = s.BpfObjCgroup.MapFloatipV4.BatchDelete(keys, &ebpf.BatchOptions{}); batchErr != nil {
// 				return fmt.Errorf("failed to BatchDelete , reason: %v ", batchErr)
// 			}
// 		}
// 		fmt.Printf("deleted item account: %v \n", count)
// 	} else {
// 		if _, batchErr := s.BpfObjCgroup.MapFloatipV4.BatchDelete(keyList, &ebpf.BatchOptions{}); batchErr != nil {
// 			return fmt.Errorf("failed to BatchDelete , reason: %v ", batchErr)
// 		}
// 	}
// 	return nil
// }
