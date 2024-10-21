// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package ebpfWriter

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	"time"
)

func (s *ebpfWriter) UpdateNode(l *zap.Logger, node *corev1.Node, onlyUpdateTime bool) error {

	if node == nil {
		return fmt.Errorf("empty node")
	}
	node.ObjectMeta.CreationTimestamp = metav1.Time{
		time.Now(),
	}

	index := node.Name
	l.Sugar().Debugf("update node %s ", index)

	entryIp, ok := node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv4]
	if ok && len(entryIp) != 0 && net.ParseIP(entryIp).To4() == nil {
		l.Sugar().Errorf("the v4 entryIp %s of node %s defined by the user is invalid ", entryIp, node.Name)
	}
	entryIp, ok = node.ObjectMeta.Annotations[types.NodeAnnotaitonNodeProxyIPv6]
	if ok && len(entryIp) != 0 && net.ParseIP(entryIp) == nil {
		l.Sugar().Errorf("the v6 entryIp %s of node %s defined by the user is invalid ", entryIp, node.Name)
		if net.ParseIP(entryIp).To4() != nil {
			l.Sugar().Errorf("the v6 entryIp %s of node %s defined by the user is not ipv6 ", entryIp, node.Name)
		}
	}

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
		l.Sugar().Infof("cache the data, and apply new data to ebpf map for the node %v , nodeIP: %+v, nodeProxyV4IP: %+v, nodeProxyV6IP: %+v", index, node.Status.Addresses, node.Annotations[types.NodeAnnotaitonNodeProxyIPv4], node.Annotations[types.NodeAnnotaitonNodeProxyIPv6])
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
