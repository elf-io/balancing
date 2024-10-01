// Code generated by bpf2go; DO NOT EDIT.

package ebpf

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/cilium/ebpf"
)

type bpf_cgroupMapkeyAffinity struct {
	ClientCookie   uint64
	OriginalDestIp uint32
	OriginalPort   uint16
	Proto          uint8
	Pad            uint8
}

type bpf_cgroupMapkeyBackend struct {
	SvcId   uint32
	Order   uint32
	Dport   uint16
	Proto   uint8
	NatType uint8
	Scope   uint8
	Pad     [3]uint8
}

type bpf_cgroupMapkeyNatRecord struct {
	SocketCookie uint64
	NatIp        uint32
	NatPort      uint16
	Proto        uint8
	Pad          uint8
}

type bpf_cgroupMapkeyService struct {
	Address uint32
	Dport   uint16
	Proto   uint8
	NatType uint8
	Scope   uint8
	Pad     [3]uint8
}

type bpf_cgroupMapvalueAffinity struct {
	Ts      uint64
	NatIp   uint32
	NatPort uint16
	Pad     [2]uint8
}

type bpf_cgroupMapvalueBackend struct {
	PodAddress uint32
	NodeId     uint32
	PodPort    uint16
	NodePort   uint16
}

type bpf_cgroupMapvalueNatRecord struct {
	OriginalDestIp   uint32
	OriginalDestPort uint16
	Pad              [2]uint8
}

type bpf_cgroupMapvalueService struct {
	SvcId             uint32
	TotalBackendCount uint32
	LocalBackendCount uint32
	AffinitySecond    uint32
	ServiceFlags      uint8
	BalancingFlags    uint8
	RedirectFlags     uint8
	NatMode           uint8
}

// loadBpf_cgroup returns the embedded CollectionSpec for bpf_cgroup.
func loadBpf_cgroup() (*ebpf.CollectionSpec, error) {
	reader := bytes.NewReader(_Bpf_cgroupBytes)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("can't load bpf_cgroup: %w", err)
	}

	return spec, err
}

// loadBpf_cgroupObjects loads bpf_cgroup and converts it into a struct.
//
// The following types are suitable as obj argument:
//
//	*bpf_cgroupObjects
//	*bpf_cgroupPrograms
//	*bpf_cgroupMaps
//
// See ebpf.CollectionSpec.LoadAndAssign documentation for details.
func loadBpf_cgroupObjects(obj interface{}, opts *ebpf.CollectionOptions) error {
	spec, err := loadBpf_cgroup()
	if err != nil {
		return err
	}

	return spec.LoadAndAssign(obj, opts)
}

// bpf_cgroupSpecs contains maps and programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpf_cgroupSpecs struct {
	bpf_cgroupProgramSpecs
	bpf_cgroupMapSpecs
}

// bpf_cgroupSpecs contains programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpf_cgroupProgramSpecs struct {
	Sock4Connect     *ebpf.ProgramSpec `ebpf:"sock4_connect"`
	Sock4Getpeername *ebpf.ProgramSpec `ebpf:"sock4_getpeername"`
	Sock4Recvmsg     *ebpf.ProgramSpec `ebpf:"sock4_recvmsg"`
	Sock4Sendmsg     *ebpf.ProgramSpec `ebpf:"sock4_sendmsg"`
}

// bpf_cgroupMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpf_cgroupMapSpecs struct {
	MapAffinity    *ebpf.MapSpec `ebpf:"map_affinity"`
	MapBackend     *ebpf.MapSpec `ebpf:"map_backend"`
	MapEvent       *ebpf.MapSpec `ebpf:"map_event"`
	MapNatRecord   *ebpf.MapSpec `ebpf:"map_nat_record"`
	MapNodeEntryIp *ebpf.MapSpec `ebpf:"map_node_entry_ip"`
	MapNodeIp      *ebpf.MapSpec `ebpf:"map_node_ip"`
	MapService     *ebpf.MapSpec `ebpf:"map_service"`
}

// bpf_cgroupObjects contains all objects after they have been loaded into the kernel.
//
// It can be passed to loadBpf_cgroupObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpf_cgroupObjects struct {
	bpf_cgroupPrograms
	bpf_cgroupMaps
}

func (o *bpf_cgroupObjects) Close() error {
	return _Bpf_cgroupClose(
		&o.bpf_cgroupPrograms,
		&o.bpf_cgroupMaps,
	)
}

// bpf_cgroupMaps contains all maps after they have been loaded into the kernel.
//
// It can be passed to loadBpf_cgroupObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpf_cgroupMaps struct {
	MapAffinity    *ebpf.Map `ebpf:"map_affinity"`
	MapBackend     *ebpf.Map `ebpf:"map_backend"`
	MapEvent       *ebpf.Map `ebpf:"map_event"`
	MapNatRecord   *ebpf.Map `ebpf:"map_nat_record"`
	MapNodeEntryIp *ebpf.Map `ebpf:"map_node_entry_ip"`
	MapNodeIp      *ebpf.Map `ebpf:"map_node_ip"`
	MapService     *ebpf.Map `ebpf:"map_service"`
}

func (m *bpf_cgroupMaps) Close() error {
	return _Bpf_cgroupClose(
		m.MapAffinity,
		m.MapBackend,
		m.MapEvent,
		m.MapNatRecord,
		m.MapNodeEntryIp,
		m.MapNodeIp,
		m.MapService,
	)
}

// bpf_cgroupPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to loadBpf_cgroupObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpf_cgroupPrograms struct {
	Sock4Connect     *ebpf.Program `ebpf:"sock4_connect"`
	Sock4Getpeername *ebpf.Program `ebpf:"sock4_getpeername"`
	Sock4Recvmsg     *ebpf.Program `ebpf:"sock4_recvmsg"`
	Sock4Sendmsg     *ebpf.Program `ebpf:"sock4_sendmsg"`
}

func (p *bpf_cgroupPrograms) Close() error {
	return _Bpf_cgroupClose(
		p.Sock4Connect,
		p.Sock4Getpeername,
		p.Sock4Recvmsg,
		p.Sock4Sendmsg,
	)
}

func _Bpf_cgroupClose(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Do not access this directly.
//
//go:embed bpf_cgroup_bpf.o
var _Bpf_cgroupBytes []byte