// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"github.com/elf-io/balancing/pkg/ebpfWriter"
	"github.com/elf-io/balancing/pkg/nodeId"
	"github.com/elf-io/balancing/pkg/types"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"reflect"
)

// -----------------------------------
type NodeReconciler struct {
	log    *zap.Logger
	writer ebpfWriter.EbpfWriter
}

func (s *NodeReconciler) HandlerAdd(obj interface{}) {
	node, ok := obj.(*corev1.Node)
	if !ok {
		s.log.Sugar().Warnf("HandlerAdd failed to get node obj: %v")
		return
	}
	logger := s.log.With(
		zap.String("node", node.Name),
	)

	logger.Sugar().Debugf("HandlerAdd process node %+v", node.Name)

	// before UpdateNode, BuildNodeId firstly
	nodeId.NodeIdManagerHander.UpdateNodeIdAndEntryIp(node)
	s.writer.UpdateNode(logger, node, false)

	// before UpdateBalancingByNode, UpdateNode firstly
	// update the nodeip and nodeProxyIp for balancing
	s.writer.UpdateBalancingByNode(logger, node)

	return
}

func checkNodeProxyIPChanged(oldNode, newNode *corev1.Node, entryKey string) bool {
	oldEntryIP, _ := oldNode.ObjectMeta.Annotations[entryKey]
	newEntryIP, _ := newNode.ObjectMeta.Annotations[entryKey]
	return oldEntryIP != newEntryIP
}

func checkNodeIdChanged(oldNode, newNode *corev1.Node) bool {
	oldId, _ := oldNode.Annotations[types.NodeAnnotationNodeIdKey]
	newId, _ := newNode.Annotations[types.NodeAnnotationNodeIdKey]
	return oldId != newId
}

func (s *NodeReconciler) HandlerUpdate(oldObj, newObj interface{}) {
	oldNode, ok1 := oldObj.(*corev1.Node)
	if !ok1 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get old node obj %v")
		return
	}
	newNode, ok2 := newObj.(*corev1.Node)
	if !ok2 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get new node obj %v")
		return
	}

	logger := s.log.With(
		zap.String("node", newNode.Name),
	)

	// update database
	nodeId.NodeIdManagerHander.UpdateNodeIdAndEntryIp(newNode)

	NoChange := true
	if t := cmp.Diff(oldNode.Status.Addresses, newNode.Status.Addresses); len(t) > 0 {
		logger.Sugar().Debugf("node address: %s", t)
	}
	if !reflect.DeepEqual(oldNode.Status.Addresses, newNode.Status.Addresses) {
		NoChange = false
		logger.Sugar().Infof("node address changed, new: %+v, old: %+v", newNode.Status.Addresses, oldNode.Status.Addresses)
	}
	if checkNodeProxyIPChanged(oldNode, newNode, types.NodeAnnotaitonNodeProxyIPv4) || checkNodeProxyIPChanged(oldNode, newNode, types.NodeAnnotaitonNodeProxyIPv6) {
		NoChange = false
		logger.Sugar().Infof("node NodeProxyIP changed, new: %+v, old: %+v", newNode.Annotations, oldNode.Annotations)
	}
	// before UpdateBalancingByNode, s.writer.UpdateNode firstly
	s.writer.UpdateNode(logger, newNode, NoChange)
	if !NoChange {
		// node ip or nodePoryIP changes, update the nodeip and nodeProxyIp for balancing
		// before UpdateBalancingByNode, s.writer.UpdateNode firstly
		s.writer.UpdateBalancingByNode(logger, newNode)
	}

	return
}

func (s *NodeReconciler) HandlerDelete(obj interface{}) {
	node, ok := obj.(*corev1.Node)
	if !ok {
		s.log.Sugar().Warnf("HandlerDelete failed to get node obj: %v")
		return
	}
	logger := s.log.With(
		zap.String("node", node.Name),
	)

	logger.Sugar().Infof("HandlerDelete process node %+v", node.Name)

	// must update the ebpf firstly, then delete the nodeIP
	nodeId.NodeIdManagerHander.DeleteNodeIdAndEntryIP(node.Name)

	// before UpdateBalancingByNode, UpdateNode firstly
	s.writer.DeleteNode(logger, node)
	// before UpdateBalancingByNode, UpdateNode firstly
	s.writer.UpdateBalancingByNode(logger, node)

	return
}

func NewNodeInformer(Client *kubernetes.Clientset, stopWatchCh chan struct{}, writer ebpfWriter.EbpfWriter) {

	// call HandlerUpdate at an interval of 60s
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(Client, InformerListInvterval)
	res := corev1.SchemeGroupVersion.WithResource("nodes")
	info, e3 := kubeInformerFactory.ForResource(res)
	if e3 != nil {
		rootLogger.Sugar().Fatalf("failed to create node informer %v", e3)
	}

	r := NodeReconciler{
		log:    rootLogger.Named("nodeReconciler"),
		writer: writer,
	}
	info.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.HandlerAdd,
		UpdateFunc: r.HandlerUpdate,
		DeleteFunc: r.HandlerDelete,
	})

	// notice that there is no need to run Start methods in a separate goroutine.
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopWatchCh)

	if !cache.WaitForCacheSync(stopWatchCh, info.Informer().HasSynced) {
		rootLogger.Sugar().Fatalf("failed to WaitForCacheSync for node ")
	}

	rootLogger.Sugar().Infof("succeeded to NewNodeInformer ")
}
