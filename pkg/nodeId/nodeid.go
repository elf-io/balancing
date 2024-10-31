// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package nodeId

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/elf-io/balancing/pkg/lock"
	"github.com/elf-io/balancing/pkg/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

/*
   each node has a persistent ID and nodeEntryIp
*/

type NodeIdManager interface {
	//
	GetNodeId(string) (uint32, error)
	GetNodeV4EntryIp(string) (string, error)
	GetNodeV6EntryIp(string) (string, error)
	//
	UpdateNodeIdAndEntryIp(*corev1.Node) (bool, bool, bool, error)
	DeleteNodeIdAndEntryIP(string)
}

type nodeIdManager struct {
	client *kubernetes.Clientset
	log    *zap.Logger

	// node id
	dataLock   *lock.Mutex
	nodeIdData map[string]uint32
	// node entry ip
	nodeIp4Data map[string]string
	nodeIp6Data map[string]string
}

// var NodeIdManagerHander NodeIdManager = (*nodeIdManager)(nil)

var _ NodeIdManager = (*nodeIdManager)(nil)

var NodeIdManagerHander NodeIdManager

// used to generate and store nodeIp for each node
// when ebpf applies some endpoints data, they need to use nodeId, but the node resource possibly has not been synchronized,
// so it introduce an abstraction layer to store and search dynamically from api-server
func InitNodeIdManager(c *kubernetes.Clientset, log *zap.Logger) {
	if _, ok := NodeIdManagerHander.(*nodeIdManager); !ok {
		t := &nodeIdManager{
			client:      c,
			nodeIdData:  make(map[string]uint32),
			nodeIp4Data: make(map[string]string),
			nodeIp6Data: make(map[string]string),
			dataLock:    &lock.Mutex{},
			log:         log,
		}
		// check node entry ip
		if len(types.AgentConfig.LocalNodeEntryInterface) > 0 {
			if ipv4, ipv6, err := CheckInterfaceAddresses(types.AgentConfig.LocalNodeEntryInterface); err != nil {
				log.Sugar().Fatalf("failed to CheckInterfaceAddresses on %s: %v", types.AgentConfig.LocalNodeEntryInterface, err)
			} else {
				if types.AgentConfig.EnableIpv4 {
					if len(ipv4) == 0 {
						log.Sugar().Fatalf("failed to get ipv4 address on entry interface %s", types.AgentConfig.LocalNodeEntryInterface)
					}
				}
				if types.AgentConfig.EnableIpv6 {
					if len(ipv6) == 0 {
						log.Sugar().Fatalf("failed to get ipv6 address on entry interface %s", types.AgentConfig.LocalNodeEntryInterface)
					}
				}
			}
		}
		//
		t.initNodeId()
		//
		NodeIdManagerHander = t
		log.Sugar().Info("finish initialize NodeIdManagerHander")
	} else {
		log.Sugar().Errorf("secondary calling for InitNodeIdManager")
	}
}

func (s *nodeIdManager) buildLocalNodeId(oldNode *corev1.Node) (nodeId uint32, ipv4Addr, ipv6Addr string, finalErr error) {

	localNodeName := oldNode.Name
	for count := 1; count < 1000; count++ {
		node, err := s.client.CoreV1().Nodes().Get(context.TODO(), localNodeName, metav1.GetOptions{})
		if err != nil {
			s.log.Sugar().Errorf("failed to get node %s: %v", localNodeName, err)
			continue
		}
		if data, ok := node.ObjectMeta.Annotations[types.NodeAnnotationNodeIdKey]; !ok {
			nodeId = generateRandomUint32()
			t := Uint32ToString(nodeId)
			node.ObjectMeta.Annotations[types.NodeAnnotationNodeIdKey] = t
			s.log.Sugar().Debugf("try to update a new nodeId %s", t)
		} else {
			if t, err1 := stringToUint32(data); err1 != nil {
				s.log.Sugar().Errorf("found an invalid nodeId %s for node %s: %v", data, node.Name, err1)
				nodeId = 0
			} else {
				nodeId = t
			}
		}
		if types.AgentConfig.EnableIpv4 {
			if data, ok := node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv4]; !ok {
				address := getNodeIpv4(node)
				if len(address) == 0 {
					s.log.Sugar().Errorf("failed to get v4 entryIp %s  ")
					ipv4Addr = ""
				} else {
					node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv4] = address
					ipv4Addr = address
					s.log.Sugar().Debugf("try to update v4 entryIp %s", address)
				}
			} else {
				if net.ParseIP(data) != nil && net.ParseIP(data).To4() != nil {
					ipv4Addr = net.ParseIP(data).To4().String()
				} else {
					s.log.Sugar().Errorf("found an invalid v4 entryIp %s for node %s", data, node.Name)
					ipv4Addr = ""
				}
			}
		}
		if types.AgentConfig.EnableIpv6 {
			if data, ok := node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv6]; !ok {
				address := getNodeIpv6(node)
				if len(address) == 0 {
					s.log.Sugar().Errorf("failed to get v6 entryIp %s  ")
					ipv6Addr = ""
				} else {
					node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv6] = address
					ipv6Addr = address
					s.log.Sugar().Debugf("try to update v6 entryIp %s", address)
				}
			} else {
				if net.ParseIP(data) != nil && net.ParseIP(data).To4() == nil {
					ipv6Addr = net.ParseIP(data).To16().String()
				} else {
					s.log.Sugar().Errorf("found an invalid v6 entryIp %s for node %s", data, node.Name)
					ipv6Addr = ""
				}
			}
		}
		// --- update
		if _, err := s.client.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{}); err != nil {
			if apierrors.IsConflict(err) {
				s.log.Sugar().Debugf("resourceVersion conflicted for node %s ", node.Name)
			} else {
				s.log.Sugar().Errorf("failed to update node %s: %v", node.Name, err)
			}
		} else {
			s.log.Sugar().Infof("succeeded to BuildLocalNodeId for node %s ", node.Name)
			return
		}
		// sleep and retry
		time.Sleep(time.Duration(count*100) * time.Millisecond)
	}

	finalErr = fmt.Errorf("failed to BuildLocalNodeId for node %s", localNodeName)
	return
}

// generate nodeId and set to the node's annotation , and build all local database
func (s *nodeIdManager) initNodeId() {

	s.log.Sugar().Infof("initial nodeId")

	// build data
	nodeList, err := s.client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.log.Sugar().Fatalf("failed to list node: %v", err)
	}

	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	// build all nodeId first
	for _, node := range nodeList.Items {
		needBuildLocalNode := false
		if nodeIdStr, ok := node.ObjectMeta.Annotations[types.NodeAnnotationNodeIdKey]; ok {
			if t, err1 := stringToUint32(nodeIdStr); err1 != nil {
				s.log.Sugar().Errorf("found an invalid nodeId %s for node %s: %v", nodeIdStr, node.Name, err1)
			} else {
				s.nodeIdData[node.Name] = t
			}
		} else {
			if node.Name == types.AgentConfig.LocalNodeName {
				needBuildLocalNode = true
			}
		}
		if types.AgentConfig.EnableIpv4 {
			if ipstr, ok := node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv4]; ok && len(ipstr) > 0 {
				if net.ParseIP(ipstr) != nil && net.ParseIP(ipstr).To4() != nil {
					s.nodeIp4Data[node.Name] = net.ParseIP(ipstr).To4().String()
				} else {
					s.log.Sugar().Errorf("found an invalid node entryIP %s for node %s: %v", ipstr, node.Name)
				}
			} else {
				if node.Name == types.AgentConfig.LocalNodeName {
					needBuildLocalNode = true
				}
			}
		}
		if types.AgentConfig.EnableIpv6 {
			if ipstr, ok := node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv6]; ok && len(ipstr) > 0 {
				if net.ParseIP(ipstr) != nil && net.ParseIP(ipstr).To4() == nil {
					s.nodeIp6Data[node.Name] = net.ParseIP(ipstr).To16().String()
				} else {
					s.log.Sugar().Errorf("found an invalid node entryIP %s for node %s: %v", ipstr, node.Name)
				}
			} else {
				if node.Name == types.AgentConfig.LocalNodeName {
					needBuildLocalNode = true
				}
			}
		}
		// ---
		if needBuildLocalNode {
			// build local node
			nodeId, ipv4Addr, ipv6Addr, err := s.buildLocalNodeId(&node)
			if err != nil {
				s.log.Sugar().Fatalf("failed to buildLocalNodeId for local node %s : %v", node.Name, err)
			} else {
				s.nodeIdData[node.Name] = nodeId
				s.nodeIp4Data[node.Name] = ipv4Addr
				s.nodeIp6Data[node.Name] = ipv6Addr
			}
		}
	}

	s.log.Sugar().Infof("succeeded to get all nodeId: %+v", s.nodeIdData)

}

func (s *nodeIdManager) UpdateNodeIdAndEntryIp(node *corev1.Node) (idChanged, ip4Chnaged, ip6Chnaged bool, err error) {
	idChanged = false
	ip4Chnaged = false
	ip6Chnaged = false

	if node == nil {
		err = fmt.Errorf("empty node obj ")
		return
	}

	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	if nodeIdStr, ok := node.ObjectMeta.Annotations[types.NodeAnnotationNodeIdKey]; ok {
		if t, err1 := stringToUint32(nodeIdStr); err1 != nil {
			s.log.Sugar().Errorf("found an invalid nodeId %s for node %s: %v", nodeIdStr, node.Name, err1)
		} else {
			if old, ok := s.nodeIdData[node.Name]; (ok && old != t) || !ok {
				s.nodeIdData[node.Name] = t
				s.log.Sugar().Infof("update nodeId from %s to %s on node %s", old, t, node.Name)
				idChanged = true
			}
		}
	}
	if types.AgentConfig.EnableIpv4 {
		if ipstr, ok := node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv4]; ok && len(ipstr) > 0 {
			if net.ParseIP(ipstr) != nil && net.ParseIP(ipstr).To4() != nil {
				if old, ok := s.nodeIp4Data[node.Name]; (old != ipstr && ok) || !ok {
					s.nodeIp4Data[node.Name] = net.ParseIP(ipstr).To4().String()
					s.log.Sugar().Infof("update nodeId from %s to %s on node %s", old, ipstr, node.Name)
					ip4Chnaged = true
				}
			} else {
				s.log.Sugar().Errorf("found an invalid node entryIP %s for node %s: %v", ipstr, node.Name)
			}
		}
	}
	if types.AgentConfig.EnableIpv6 {
		if ipstr, ok := node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv6]; ok && len(ipstr) > 0 {
			if net.ParseIP(ipstr) != nil && net.ParseIP(ipstr).To4() == nil {
				if old, ok := s.nodeIp6Data[node.Name]; (old != ipstr && ok) || !ok {
					s.nodeIp6Data[node.Name] = net.ParseIP(ipstr).To16().String()
					s.log.Sugar().Infof("update nodeId from %s to %s on node %s", old, ipstr, node.Name)
					ip6Chnaged = true
				}
			} else {
				s.log.Sugar().Errorf("found an invalid node entryIP %s for node %s: %v", ipstr, node.Name)
			}
		}
	}

	return
}

func (s *nodeIdManager) GetNodeId(nodeName string) (uint32, error) {
	if len(nodeName) == 0 {
		s.log.Sugar().Errorf("empty nodeName ")
		return 0, fmt.Errorf("empty nodeName ")
	}

	// for redirect and balancing polity, it does not care about nodeId
	if nodeName == types.NodeNameIgnore {
		return 0xffffffff, nil
	}

	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	if nodeId, ok := s.nodeIdData[nodeName]; ok {
		return nodeId, nil
	}
	s.log.Sugar().Errorf("no node Id for node %s ", nodeName)

	return 0, fmt.Errorf("no node Id")
}

func (s *nodeIdManager) GetNodeV4EntryIp(nodeName string) (string, error) {
	if len(nodeName) == 0 {
		s.log.Sugar().Errorf("empty nodeName ")
		return "", fmt.Errorf("empty nodeName ")
	}

	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	if data, ok := s.nodeIp4Data[nodeName]; ok {
		return data, nil
	}
	s.log.Sugar().Errorf("no node V4EntryIp for node %s ", nodeName)

	return "", fmt.Errorf("no node V4EntryIp")
}

func (s *nodeIdManager) GetNodeV6EntryIp(nodeName string) (string, error) {
	if len(nodeName) == 0 {
		s.log.Sugar().Errorf("empty nodeName ")
		return "", fmt.Errorf("empty nodeName ")
	}

	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	if data, ok := s.nodeIp6Data[nodeName]; ok {
		return data, nil
	}
	s.log.Sugar().Errorf("no node V6EntryIp for node %s ", nodeName)

	return "", fmt.Errorf("no node V6EntryIp")
}

func (s *nodeIdManager) DeleteNodeIdAndEntryIP(nodeName string) {
	if len(nodeName) == 0 {
		return
	}
	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	delete(s.nodeIdData, nodeName)
	delete(s.nodeIp4Data, nodeName)
	delete(s.nodeIp6Data, nodeName)

	s.log.Sugar().Infof("succeeded to delete nodeId for node %s", nodeName)
}
