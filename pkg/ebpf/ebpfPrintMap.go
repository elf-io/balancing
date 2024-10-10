package ebpf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/perf"
)

// PrintMapService prints the map data for services, categorized by NatType.
func (s *EbpfProgramStruct) PrintMapService() error {

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

	// Temporary storage for categorized keys and values
	var allKeysRedirect []bpf_cgroupMapkeyService
	var allValsRedirect []bpf_cgroupMapvalueService
	var allKeysService []bpf_cgroupMapkeyService
	var allValsService []bpf_cgroupMapvalueService
	var allKeysBalancing []bpf_cgroupMapkeyService
	var allValsBalancing []bpf_cgroupMapvalueService

	for {
		keys := make([]bpf_cgroupMapkeyService, 100)
		vals := make([]bpf_cgroupMapvalueService, 100)

		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				finished = true
			} else {
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		// Categorize current batch
		for i := 0; i < c; i++ {
			switch keys[i].NatType {
			case NAT_TYPE_REDIRECT:
				allKeysRedirect = append(allKeysRedirect, keys[i])
				allValsRedirect = append(allValsRedirect, vals[i])
			case NAT_TYPE_SERVICE:
				allKeysService = append(allKeysService, keys[i])
				allValsService = append(allValsService, vals[i])
			case NAT_TYPE_BALANCING:
				allKeysBalancing = append(allKeysBalancing, keys[i])
				allValsBalancing = append(allValsBalancing, vals[i])
			default:
				fmt.Printf("Unknown NatType: %v : key=%s, value=%s\n", i, keys[i], vals[i])
			}
		}
		if finished {
			break
		}
	}

	// Print categorized data
	fmt.Println("")
	fmt.Println("Service Entries: ", len(allKeysService))
	for i := 0; i < len(allKeysService); i++ {
		fmt.Printf("[%v]: key=%s, \n", i, allKeysService[i])
		fmt.Printf("     value=%s\n", allValsService[i])
	}

	fmt.Println("")
	fmt.Println("LocalRedirect Entries: ", len(allKeysRedirect))
	for i := 0; i < len(allKeysRedirect); i++ {
		fmt.Printf("[%v]: key=%s\n", i, allKeysRedirect[i])
		fmt.Printf("     value=%s\n", allValsRedirect[i])
	}

	fmt.Println("")
	fmt.Println("Balancing Entries: ", len(allKeysBalancing))
	for i := 0; i < len(allKeysBalancing); i++ {
		fmt.Printf("[%v]: key=%s\n", i, allKeysBalancing[i])
		fmt.Printf("     value=%s\n", allValsBalancing[i])
	}

	fmt.Println("")
	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

// PrintMapBackend prints the map data for backends, categorized by NatType.
func (s *EbpfProgramStruct) PrintMapBackend() error {
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

	// Temporary storage for categorized keys and values
	var allKeysRedirect []bpf_cgroupMapkeyBackend
	var allValsRedirect []bpf_cgroupMapvalueBackend
	var allKeysService []bpf_cgroupMapkeyBackend
	var allValsService []bpf_cgroupMapvalueBackend
	var allKeysBalancing []bpf_cgroupMapkeyBackend
	var allValsBalancing []bpf_cgroupMapvalueBackend

	for {
		keys := make([]bpf_cgroupMapkeyBackend, 100)
		vals := make([]bpf_cgroupMapvalueBackend, 100)

		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				finished = true
			} else {
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		// Categorize current batch
		for i := 0; i < c; i++ {
			switch keys[i].NatType {
			case NAT_TYPE_REDIRECT:
				allKeysRedirect = append(allKeysRedirect, keys[i])
				allValsRedirect = append(allValsRedirect, vals[i])
			case NAT_TYPE_SERVICE:
				allKeysService = append(allKeysService, keys[i])
				allValsService = append(allValsService, vals[i])
			case NAT_TYPE_BALANCING:
				allKeysBalancing = append(allKeysBalancing, keys[i])
				allValsBalancing = append(allValsBalancing, vals[i])
			default:
				fmt.Printf("Unknown NatType: %v : key=%s, value=%s\n", i, keys[i], vals[i])
			}
		}
		if finished {
			break
		}
	}

	// Print categorized data
	fmt.Println("")
	fmt.Println("Service Entries: ", len(allKeysService))
	for i := 0; i < len(allKeysService); i++ {
		fmt.Printf("[%v]: key=%s\n", i, allKeysService[i])
		fmt.Printf("     value=%s\n", allValsService[i])
	}

	fmt.Println("")
	fmt.Println("LocalRedirect Entries : ", len(allKeysRedirect))
	for i := 0; i < len(allKeysRedirect); i++ {
		fmt.Printf("[%v]: key=%s\n", i, allKeysRedirect[i])
		fmt.Printf("     value=%s\n", allValsRedirect[i])
	}

	fmt.Println("")
	fmt.Println("Balancing Entries: ", len(allKeysBalancing))
	for i := 0; i < len(allKeysBalancing); i++ {
		fmt.Printf("[%v]: key=%s\n", i, allKeysBalancing[i])
		fmt.Printf("     value=%s\n", allValsBalancing[i])
	}

	fmt.Println("")
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
			fmt.Printf("[%v]: key=%s\n", i, keys[i])
			fmt.Printf("      value=%+v\n", vals[i])
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
			fmt.Printf("[%v]: key=%+v\n", i, keys[i])
			fmt.Printf("      value=%s\n", vals[i])
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
			fmt.Printf("[%v]: key=%s\n", i, keys[i])
			fmt.Printf("      value=%s\n", vals[i])
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
			fmt.Printf("[%v]: key=%s\n", i, keys[i])
			fmt.Printf("      value=%s\n", vals[i])
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
