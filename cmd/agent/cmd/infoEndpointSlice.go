// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpfWriter"
	"github.com/elf-io/balancing/pkg/k8s"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	discoveryv1 "k8s.io/api/discovery/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"reflect"
)

// -----------------------------------
type EndpoingSliceReconciler struct {
	log    *zap.Logger
	writer ebpfWriter.EbpfWriter
}

func (s *EndpoingSliceReconciler) HandlerAdd(obj interface{}) {
	eds, ok := obj.(*discoveryv1.EndpointSlice)
	if !ok {
		s.log.Sugar().Warnf("HandlerAdd failed to get EndpointSlice obj: %v")
		return
	}

	name := eds.Namespace + "/" + eds.Name
	logger := s.log.With(
		zap.String("loadbalance", k8s.GetEndpointSliceOwnerName(eds)),
		zap.String("endpointslice", name),
	)

	logger.Sugar().Infof("HandlerAdd process EndpointSlice: %+v", name)
	if err := s.writer.UpdateServiceByEndpointSlice(logger, eds, false); err != nil {
		// 处理错误
		fmt.Println("Error:", err)
	}

}

func (s *EndpoingSliceReconciler) HandlerUpdate(oldObj, newObj interface{}) {
	oldEds, ok1 := oldObj.(*discoveryv1.EndpointSlice)
	if !ok1 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get old EndpointSlice obj: %v")
		return
	}
	newEds, ok2 := newObj.(*discoveryv1.EndpointSlice)
	if !ok2 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get new EndpointSlice obj: %v")
		return
	}

	name := newEds.Namespace + "/" + newEds.Name
	logger := s.log.With(
		zap.String("loadbalance", k8s.GetEndpointSliceOwnerName(newEds)),
		zap.String("endpointslice", name),
	)

	NoChange := false
	if t := cmp.Diff(oldEds, newEds); len(t) > 0 {
		logger.Sugar().Debugf("EndpointSlice diff: %s", t)
	}
	if reflect.DeepEqual(oldEds.Endpoints, newEds.Endpoints) && reflect.DeepEqual(oldEds.Ports, newEds.Ports) {
		NoChange = true
	}
	if !NoChange {
		logger.Sugar().Infof("HandlerUpdate process changed EndpointSlice %+v", name)
	} else {
		logger.Sugar().Debugf("HandlerUpdate process EndpointSlice %+v", name)
	}

	logger.Sugar().Infof("HandlerUpdate process EndpointSlice: %+s", name)
	if err := s.writer.UpdateServiceByEndpointSlice(logger, newEds, NoChange); err != nil {
		// 处理错误
		fmt.Println("Error:", err)
	}

}

func (s *EndpoingSliceReconciler) HandlerDelete(obj interface{}) {
	eds, ok := obj.(*discoveryv1.EndpointSlice)
	if !ok {
		s.log.Sugar().Warnf("HandlerDelete failed to get EndpointSlice obj: %v")
		return
	}
	name := eds.Namespace + "/" + eds.Name
	logger := s.log.With(
		zap.String("loadbalance", k8s.GetEndpointSliceOwnerName(eds)),
		zap.String("endpointslice", name),
	)

	logger.Sugar().Infof("HandlerDelete process EndpointSlice: %s", name)
	if err := s.writer.DeleteServiceByEndpointSlice(logger, eds); err != nil {
		// 处理错误
		fmt.Println("Error:", err)
	}

}

func NewEndpointSliceInformer(Client *kubernetes.Clientset, stopWatchCh chan struct{}, writer ebpfWriter.EbpfWriter) {

	// call HandlerUpdate at an interval of 60s
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(Client, InformerListInvterval)
	// service
	edsRes := discoveryv1.SchemeGroupVersion.WithResource("endpointslices")
	srcInformer, e3 := kubeInformerFactory.ForResource(edsRes)
	if e3 != nil {
		rootLogger.Sugar().Fatalf("failed to create service informer: %v", e3)
	}

	r := EndpoingSliceReconciler{
		log:    rootLogger.Named("EndpointsliceReconciler"),
		writer: writer,
	}
	t := cache.ResourceEventHandlerFuncs{
		AddFunc:    r.HandlerAdd,
		UpdateFunc: r.HandlerUpdate,
		DeleteFunc: r.HandlerDelete,
	}
	if _, e := srcInformer.Informer().AddEventHandler(t); e != nil {
		rootLogger.Sugar().Fatalf("failed to AddEventHandler: %v", e)

	}

	// notice that there is no need to run Start methods in a separate goroutine.
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopWatchCh)

	if !cache.WaitForCacheSync(stopWatchCh, srcInformer.Informer().HasSynced) {
		rootLogger.Sugar().Fatalf("failed to WaitForCacheSync for endpointslice ")
	}
	rootLogger.Sugar().Infof("succeeded to NewEndpointSliceInformer")

}
