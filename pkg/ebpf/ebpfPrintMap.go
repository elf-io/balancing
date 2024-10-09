package ebpf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/perf"
	"os"
)

// func (s *EbpfProgramStruct) PrintMapDataService() error {
//
// 	keys := make([]bpf_cgroupMapkeyService, 100)
// 	vals := make([]bpf_cgroupMapvalueService, 100)
//
// 	var cursor ebpf.MapBatchCursor
// 	count := 0
// 	for {
// 		c, batchErr := s.BpfObjCgroup.MapService.BatchLookup(&cursor, keys, vals, nil)
// 		count += c
// 		finished := false
// 		if batchErr != nil {
// 			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
// 				// end
// 				finished = true
// 			} else {
// 				return fmt.Errorf("failed to batchlookup for %v\n", s.BpfObjCgroup.MapService.String())
// 			}
// 		}
// 		for i := 0; i < len(keys) && i < c; i++ {
// 			fmt.Printf("%v : key=%+v\n", i, keys[i])
// 			fmt.Printf("%v : value=%+v\n", i, vals[i])
// 			fmt.Printf("\n")
// 		}
// 		if finished {
// 			break
// 		}
// 	}
// 	fmt.Printf("end map %s: total items: %v \n",name, count)
// 	return nil
// }

// type MapName string
// const (
// 	MapNameService   = MapName("mapService")
// 	MapNameAffinity  = MapName("mapAffinity")
// 	MapNameBackend   = MapName("mapBackend")
// 	MapNameNatRecord = MapName("mapNatRecord")
// 	MapNameNode      = MapName("mapNode")
// )

func (s *EbpfProgramStruct) PrintMapService() error {
	keys := make([]bpf_cgroupMapkeyService, 100)
	vals := make([]bpf_cgroupMapvalueService, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapService != nil {
		mapPtr = s.BpfObjCgroup.MapService
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapService != nil {
		mapPtr = s.EbpfMaps.MapService
	} else {
		return fmt.Errorf("failed to get ebpf map")
	}
	name := mapPtr.String()

	fmt.Printf("------------------------------\n")
	fmt.Printf("map  %s\n", name)
	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c; i++ {
			fmt.Printf("%v : key=%+v\n", i, keys[i])
			fmt.Printf("%v : value=%+v\n", i, vals[i])
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

func (s *EbpfProgramStruct) PrintMapBackend() error {
	keys := make([]bpf_cgroupMapkeyBackend, 100)
	vals := make([]bpf_cgroupMapvalueBackend, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapBackend != nil {
		mapPtr = s.BpfObjCgroup.MapBackend
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapBackend != nil {
		mapPtr = s.EbpfMaps.MapBackend
	} else {
		return fmt.Errorf("failed to get ebpf map")
	}
	name := mapPtr.String()

	fmt.Printf("------------------------------\n")
	fmt.Printf("map  %s\n", name)
	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c; i++ {
			fmt.Printf("%v : key=%+v\n", i, keys[i])
			fmt.Printf("%v : value=%+v\n", i, vals[i])
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

func (s *EbpfProgramStruct) PrintMapNodeIp() error {
	keys := make([]bpf_cgroupMapkeyNodeIp, 100)
	vals := make([]uint32, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapNodeIp != nil {
		mapPtr = s.BpfObjCgroup.MapNodeIp
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapNodeIp != nil {
		mapPtr = s.EbpfMaps.MapNodeIp
	} else {
		return fmt.Errorf("failed to get ebpf map")
	}
	name := mapPtr.String()

	fmt.Printf("------------------------------\n")
	fmt.Printf("map  %s\n", name)
	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c; i++ {
			fmt.Printf("%v : key=%+v\n", i, keys[i])
			fmt.Printf("%v : value=%+v\n", i, vals[i])
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

func (s *EbpfProgramStruct) PrintMapNodeProxyIp() error {
	keys := make([]uint32, 100)
	vals := make([]bpf_cgroupMapvalueNodeProxyIp, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapNodeProxyIp != nil {
		mapPtr = s.BpfObjCgroup.MapNodeProxyIp
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapNodeProxyIp != nil {
		mapPtr = s.EbpfMaps.MapNodeProxyIp
	} else {
		return fmt.Errorf("failed to get ebpf map")
	}
	name := mapPtr.String()

	fmt.Printf("------------------------------\n")
	fmt.Printf("map  %s\n", name)
	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c; i++ {
			fmt.Printf("%v : key=%+v\n", i, keys[i])
			fmt.Printf("%v : value=%+v\n", i, vals[i])
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

func (s *EbpfProgramStruct) PrintMapAffinity() error {
	keys := make([]bpf_cgroupMapkeyAffinity, 100)
	vals := make([]bpf_cgroupMapvalueAffinity, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapAffinity != nil {
		mapPtr = s.BpfObjCgroup.MapAffinity
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapAffinity != nil {
		mapPtr = s.EbpfMaps.MapAffinity
	} else {
		return fmt.Errorf("failed to get ebpf map")
	}
	name := mapPtr.String()

	fmt.Printf("------------------------------\n")
	fmt.Printf("map  %s\n", name)
	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c; i++ {
			fmt.Printf("%v : key=%+v\n", i, keys[i])
			fmt.Printf("%v : value=%+v\n", i, vals[i])
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

func (s *EbpfProgramStruct) PrintMapNatRecord() error {
	keys := make([]bpf_cgroupMapkeyNatRecord, 100)
	vals := make([]bpf_cgroupMapvalueNatRecord, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapNatRecord != nil {
		mapPtr = s.BpfObjCgroup.MapNatRecord
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapNatRecord != nil {
		mapPtr = s.EbpfMaps.MapNatRecord
	} else {
		return fmt.Errorf("failed to get ebpf map")
	}
	name := mapPtr.String()

	fmt.Printf("------------------------------\n")
	fmt.Printf("map  %s\n", name)
	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c; i++ {
			fmt.Printf("%v : key=%+v\n", i, keys[i])
			fmt.Printf("%v : value=%+v\n", i, vals[i])
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

func (s *EbpfProgramStruct) PrintMapConfigure() error {
	keys := make([]uint32, 100)
	vals := make([]uint32, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapConfigure != nil {
		mapPtr = s.BpfObjCgroup.MapConfigure
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapConfigure != nil {
		mapPtr = s.EbpfMaps.MapConfigure
	} else {
		return fmt.Errorf("failed to get ebpf map")
	}
	name := mapPtr.String()

	fmt.Printf("------------------------------\n")
	fmt.Printf("map  %s\n", name)

	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c && i < MapConfigureKeyIndexEnd; i++ {
			fmt.Printf("%s\n", MapConfigureStr(uint32(i), uint32(vals[i])))
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

// -------------------------- event map

func (s *EbpfProgramStruct) GetMapDataEvent() <-chan MapEventValue {
	return s.Event

}

// get data from map
func (s *EbpfProgramStruct) daemonGetEvent() {

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapEvent != nil {
		mapPtr = s.BpfObjCgroup.MapEvent
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapEvent != nil {
		mapPtr = s.EbpfMaps.MapEvent
	} else {
		s.l.Sugar().Fatal("failed to get ebpf event map")
	}

	rd, err := perf.NewReader(mapPtr, os.Getpagesize())
	if err != nil {
		s.l.Sugar().Fatal("failed to read ebpf map : %v ", err)
	}
	defer rd.Close()

	for {
		record, err := rd.Read()
		if err != nil {
			s.l.Sugar().Warnf("failed to read event: %v", err)
			continue
		}

		t := MapEventValue{}
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.NativeEndian, &t); err != nil {
			s.l.Sugar().Warnf("parsing ringbuf event: %s", err)
			continue
		}
		s.l.Sugar().Debugf("raw ebpf event: %s ", t)

		select {
		case s.Event <- t:
		default:
			s.l.Sugar().Warnf("failed to write data to event chan, miss data: %v \n", t)
		}
	}
}
