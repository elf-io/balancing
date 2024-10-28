package ebpfEvent

import (
	"fmt"
	"time"

	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/elf-io/balancing/pkg/ebpfWriter"
	"github.com/elf-io/balancing/pkg/podId"
	"github.com/elf-io/balancing/pkg/types"
	"go.uber.org/zap"
)

type ebpfEventStruct struct {
	l           *zap.Logger
	ebpfHandler ebpf.EbpfProgram
	writer      ebpfWriter.EbpfWriter
}

type EbpfEvent interface {
	WatchEbpfEvent(chan struct{})
}

var _ EbpfEvent = (*ebpfEventStruct)(nil)

func NewEbpfEvent(l *zap.Logger, ebpfHandler ebpf.EbpfProgram, writer ebpfWriter.EbpfWriter) EbpfEvent {
	return &ebpfEventStruct{
		l:           l,
		ebpfHandler: ebpfHandler,
		writer:      writer,
	}
}

// 定义一个结构体来存储事件信息
type EventInfo struct {
	ClientPodName     string
	ClientNamespace   string
	ClientPid         string
	ClientPodUuid     string
	ClientContainerId string
	IsHostApp         bool
	NodeName          string
	IsIpv4            bool
	IsSuccess         bool
	DestIp            string
	DestPort          string
	NatIp             string
	NatPort           string
	Failure           string
	TimeStamp         string
	ServiceId         string
	PolicyName        string
	NatType           string
	NatMode           string
	Protocol          string
}

func (s *ebpfEventStruct) WatchEbpfEvent(stopWatch chan struct{}) {
	var err error
	go func() {
		eventCh := s.ebpfHandler.GetMapDataEvent()

		for {
			select {
			case <-stopWatch:
				s.l.Sugar().Infof("stop watch ebpf event")
				break
			case event, ok := <-eventCh:
				if !ok {
					s.l.Sugar().Fatalf("ebpf event channel is closed")
				}

				s.l.Sugar().Debugf("received an ebpf event: %s ", event)

				eventInfo := EventInfo{
					NodeName:  types.AgentConfig.LocalNodeName,
					IsIpv4:    event.IsIpv4 != 0,
					IsSuccess: event.IsSuccess != 0,
					ClientPid: fmt.Sprintf("%d", event.Pid),
					Failure:   ebpf.GetFailureStr(event.FailureCode),
					TimeStamp: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
					ServiceId: fmt.Sprintf("%d", event.SvcId),
					NatType:   ebpf.GetNatTypeStr(event.NatType),
					NatMode:   ebpf.GetNatModeStr(event.NatMode),
					Protocol:  ebpf.GetProtocolStr(event.Protocol),
				}

				eventInfo.ClientPodName, eventInfo.ClientNamespace, eventInfo.ClientContainerId, eventInfo.ClientPodUuid, eventInfo.IsHostApp, err = podId.PodIdHander.LookupPodByPid(event.Pid)
				if err != nil {
					s.l.Sugar().Errorf("failed to get podName for pid %d: %v", event.Pid, err)
				}

				if event.IsIpv4 != 0 {
					eventInfo.DestIp = ebpf.GetIpStr(event.OriginalDestV4Ip)
					eventInfo.DestPort = fmt.Sprintf("%d", event.OriginalDestPort)
					eventInfo.NatIp = ebpf.GetIpStr(event.NatV4Ip)
					eventInfo.NatPort = fmt.Sprintf("%d", event.NatPort)
				} else {
					eventInfo.DestIp = ebpf.GetIpv6Str(event.OriginalDestV6ipHigh, event.OriginalDestV6ipLow)
					eventInfo.DestPort = fmt.Sprintf("%d", event.OriginalDestPort)
					eventInfo.NatIp = ebpf.GetIpv6Str(event.NatV6ipHigh, event.NatV6ipLow)
					eventInfo.NatPort = fmt.Sprintf("%d", event.NatPort)
				}

				namespace, name, err := s.writer.GetPolicyBySvcId(event.NatType, event.SvcId)
				if err != nil {
					s.l.Sugar().Errorf("failed to find policy for ebpf event with natMode=%s and svcId=%d : %v ", ebpf.GetNatTypeStr(event.NatType), event.SvcId, err)
					eventInfo.PolicyName = ""
				} else {
					switch event.NatType {
					case ebpf.NAT_TYPE_SERVICE:
						eventInfo.PolicyName = fmt.Sprintf("%s/%s", namespace, name)
					default:
						eventInfo.PolicyName = name
					}
				}

				s.l.Sugar().Infof("formatted ebpf event: %+v", eventInfo)
			}
		}
	}()
}
