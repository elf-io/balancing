// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package nodeIp

import (
	"github.com/elf-io/balancing/pkg/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"net"
)

type NodeEntryIPManager interface {
	UpdateNodeEntryIP(node *corev1.Node)
}

type nodeEntryIPManager struct {
	l               *zap.Logger
	client          *kubernetes.Clientset
	entryiInterface string
}

var _ NodeEntryIPManager = (*nodeEntryIPManager)nil
var NodeEntryIPManagerHander NodeEntryIPManager


func NewNodeEntryIPManager(c *kubernetes.Clientset, log *zap.Logger) {
	if _, ok := NodeEntryIPManagerHander.(*nodeEntryIPManager); !ok {
		t := &nodeEntryIPManager{
			l:               log,
			client:          c,
			entryiInterface: types.AgentConfig.LocalNodeEntryInterface,
		}
		log.Sugar().Info("finish initialize NewNodeEntryIPManager")

	} else {
		log.Sugar().Errorf("secondary calling for NewNodeEntryIPManager")
	}
}

func getNodeIpv4() {

}

func getNodeIpv6() {

}

func (s *nodeEntryIPManager) UpdateNodeEntryIP(node *corev1.Node) {

	entryIp, ok := node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv4]
	if ok && len(entryIp) != 0 && net.ParseIP(entryIp).To4() == nil {
		s.l.Sugar().Errorf("the v4 entryIp %s of node %s defined by the user is invalid ", entryIp, node.Name)
	}

	entryIp, ok = node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv6]
	if ok && len(entryIp) != 0 && net.ParseIP(entryIp) == nil {
		s.l.Sugar().Errorf("the v6 entryIp %s of node %s defined by the user is invalid ", entryIp, node.Name)
		if net.ParseIP(entryIp).To4() != nil {
			s.l.Sugar().Errorf("the v6 entryIp %s of node %s defined by the user is not ipv6 ", entryIp, node.Name)
		}
	}
}
