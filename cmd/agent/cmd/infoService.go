// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"github.com/elf-io/balancing/pkg/ebpfWriter"
	"github.com/elf-io/balancing/pkg/utils"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"reflect"
)

// -----------------------------------
type ServiceReconciler struct {
	log    *zap.Logger
	writer ebpfWriter.EbpfWriter
}

func SkipServiceProcess(svc *corev1.Service) bool {
	switch svc.Spec.Type {
	case corev1.ServiceTypeClusterIP:
		return false
	case corev1.ServiceTypeNodePort:
		return false
	case corev1.ServiceTypeLoadBalancer:
		return false
	}
	return true
}

func (s *ServiceReconciler) HandlerAdd(obj interface{}) {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		s.log.Sugar().Warnf("HandlerAdd failed to get service obj: %v")
		return
	}
	name := svc.Namespace + "/" + svc.Name
	logger := s.log.With(
		zap.String("loadbalance", name),
		zap.String("service", name),
	)

	if SkipServiceProcess(svc) {
		logger.Sugar().Debugf("HandlerAdd skip unsupported service %+v", name)
		return
	}

	logger.Sugar().Infof("HandlerAdd process service %+v", name)
	// update service
	if true {
		newSvc := corev1.Service{}
		if e := utils.DeepCopy(svc, &newSvc); e != nil {
			logger.Sugar().Errorf("failed to DeepCopy service: %+v", e)
			return
		}
		// use a copied service in case of  modification
		s.writer.UpdateServiceByService(logger, &newSvc, false)
	}
	// update localRedirect
	if true {
		newSvc := corev1.Service{}
		if e := utils.DeepCopy(svc, &newSvc); e != nil {
			logger.Sugar().Errorf("failed to DeepCopy service: %+v", e)
			return
		}
		// use a copied service in case of  modification
		s.writer.UpdateRedirectByService(logger, &newSvc)
	}
	// update balancing
	if true {
		newSvc := corev1.Service{}
		if e := utils.DeepCopy(svc, &newSvc); e != nil {
			logger.Sugar().Errorf("failed to DeepCopy service: %+v", e)
			return
		}
		// use a copied service in case of  modification
		s.writer.UpdateBalancingByService(logger, &newSvc)
	}

	return
}

func (s *ServiceReconciler) HandlerUpdate(oldObj, newObj interface{}) {
	oldSvc, ok1 := oldObj.(*corev1.Service)
	if !ok1 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get old service obj %v")
		return
	}
	newSvc, ok2 := newObj.(*corev1.Service)
	if !ok2 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get new service obj %v")
		return
	}

	name := newSvc.Namespace + "/" + newSvc.Name
	logger := s.log.With(
		zap.String("loadbalance", name),
		zap.String("service", name),
	)

	if SkipServiceProcess(newSvc) && SkipServiceProcess(oldSvc) {
		logger.Sugar().Debugf("HandlerAdd skip unsupported service %+v", name)
		return
	}

	// update service
	NoChange := false
	if t := cmp.Diff(oldSvc, newSvc); len(t) > 0 {
		logger.Sugar().Debugf("service diff: %s", t)
	}
	if reflect.DeepEqual(oldSvc.Spec, newSvc.Spec) && reflect.DeepEqual(oldSvc.Status, newSvc.Status) {
		NoChange = true
	}
	if !NoChange {
		logger.Sugar().Infof("HandlerUpdate process changed service %+v", name)
	} else {
		logger.Sugar().Debugf("HandlerUpdate process service %+v", name)
	}
	if true {
		svc := corev1.Service{}
		if e := utils.DeepCopy(newSvc, &svc); e != nil {
			logger.Sugar().Errorf("failed to DeepCopy service: %+v", e)
			return
		}
		// use a copied service in case of  modification
		s.writer.UpdateServiceByService(logger, &svc, NoChange)
	}

	// update localRedirect
	if !NoChange {
		svc := corev1.Service{}
		if e := utils.DeepCopy(newSvc, &svc); e != nil {
			logger.Sugar().Errorf("failed to DeepCopy service: %+v", e)
			return
		}
		// use a copied service in case of  modification
		s.writer.UpdateRedirectByService(logger, &svc)
	}
	// update balancing
	if !NoChange {
		svc := corev1.Service{}
		if e := utils.DeepCopy(newSvc, &svc); e != nil {
			logger.Sugar().Errorf("failed to DeepCopy service: %+v", e)
			return
		}
		// use a copied service in case of  modification
		s.writer.UpdateBalancingByService(logger, &svc)
	}

	return
}

func (s *ServiceReconciler) HandlerDelete(obj interface{}) {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		s.log.Sugar().Warnf("HandlerDelete failed to get service obj: %v")
		return
	}
	name := svc.Namespace + "/" + svc.Name
	logger := s.log.With(
		zap.String("loadbalance", name),
		zap.String("service", name),
	)

	if SkipServiceProcess(svc) {
		logger.Sugar().Debugf("HandlerAdd skip service %+v", name)
		return
	}
	logger.Sugar().Infof("HandlerDelete process service %+v", svc)
	// update service
	s.writer.DeleteServiceByService(logger, svc)

	// update localRedirect
	s.writer.DeleteRedirectByService(logger, svc)
	// update balancing
	s.writer.DeleteBalancingByService(logger, svc)

	return
}

func NewServiceInformer(Client *kubernetes.Clientset, stopWatchCh chan struct{}, writer ebpfWriter.EbpfWriter) {

	// call HandlerUpdate at an interval of 60s
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(Client, InformerListInvterval)
	// service
	svcRes := corev1.SchemeGroupVersion.WithResource("services")
	srcInformer, e3 := kubeInformerFactory.ForResource(svcRes)
	if e3 != nil {
		rootLogger.Sugar().Fatalf("failed to create service informer %v", e3)
	}

	r := ServiceReconciler{
		log:    rootLogger.Named("serviceReconciler"),
		writer: writer,
	}
	srcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.HandlerAdd,
		UpdateFunc: r.HandlerUpdate,
		DeleteFunc: r.HandlerDelete,
	})

	// notice that there is no need to run Start methods in a separate goroutine.
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopWatchCh)

	if !cache.WaitForCacheSync(stopWatchCh, srcInformer.Informer().HasSynced) {
		rootLogger.Sugar().Fatalf("failed to WaitForCacheSync for service ")
	}

	rootLogger.Sugar().Infof("succeeded to NewServiceInformer ")
}
