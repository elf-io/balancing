// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package mybookManager

import (
	"context"
	"fmt"
	"github.com/elf-io/elf/pkg/k8s"
	crd "github.com/elf-io/elf/pkg/k8s/apis/elf.elf.io/v1"
	crdclientset "github.com/elf-io/elf/pkg/k8s/client/clientset/versioned"
	"github.com/elf-io/elf/pkg/k8s/client/informers/externalversions"
	crdlisterv1 "github.com/elf-io/elf/pkg/k8s/client/listers/elf.elf.io/v1"
	"github.com/elf-io/elf/pkg/lease"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const (
	workerNumber    = 2
	queueMaxRetries = 100
)

type informerHandler struct {
	logger         *zap.Logger
	leaseName      string
	leaseNameSpace string
	leaseId        string
	eventRecord    record.EventRecorder
	queue          workqueue.RateLimitingInterface
	crdlister      crdlisterv1.MybookLister
	k8sclient      crdclientset.Interface
}

func (s *informerHandler) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

func (s *informerHandler) processNextWorkItem(ctx context.Context) bool {
	key, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(key)

	err := s.syncHandler(ctx, key.(*crd.Mybook))
	if err == nil {
		s.queue.Forget(key)
	} else {
		s.logger.Sugar().Warnf("worker failed , error=%v", err)
		if apierrors.IsConflict(err) {
			// 更新CRD 时，resourceVersion 冲突，重试
			s.queue.AddRateLimited(key)
		} else if s.queue.NumRequeues(key) < queueMaxRetries {
			s.queue.AddRateLimited(key)
		}
	}
	// handle nex item
	return true
}

// ===================================

func (s *informerHandler) informerAddHandler(obj interface{}) {

	r, ok := obj.(*crd.Mybook)
	if !ok {
		s.logger.Sugar().Errorf("failed to get crd: %+v", obj)
		return
	}
	s.logger.Sugar().Infof("add mybook : %+v", r)

	// enqueue
	s.queue.AddRateLimited(r)

	// 基于工作队列，生产者和消费者模型，使得事件能够本可靠的 异步执行 成功
	// 例如 resourceVersion、断网等失败，最终都能够不断重试而手工
	s.eventRecord.Eventf(r, corev1.EventTypeNormal, "newMybook", "crd event, new mybook %v", r.Name)

}

func (s *informerHandler) informerUpdateHandler(oldObj interface{}, newObj interface{}) {
	_, ok := oldObj.(*crd.Mybook)
	if !ok {
		s.logger.Sugar().Errorf("failed to get crd: %+v", oldObj)
		return
	}
	newone, ok := newObj.(*crd.Mybook)
	if !ok {
		s.logger.Sugar().Errorf("failed to get crd: %+v", newObj)
		return
	}
	s.logger.Sugar().Infof("update crd: %v", newone.Name)

	// // 简单方式处理事件，堵塞执行，重建重试5次更新，
	// // 好处是代码简单，坏处是，resourceVersion、断网等场景，可能最终5次失败而最终失败
	// if !reflect.DeepEqual(curPod.Spec, oldPod.Spec) {
	// 	// when errors.IsConflict owing to resourceVersion, auto retry 5 times at interval 10ms
	// 	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
	// 		return s.handleCrdEvent(context.Background(), curPod)
	// 	})
	// 	if err != nil {
	// 		s.logger.Sugar().Errorf("failed to update mybook status, error=%v", err)
	// 	}
	// }
}

func (s *informerHandler) informerDeleteHandler(obj interface{}) {
	r, ok := obj.(*crd.Mybook)
	if !ok {
		s.logger.Sugar().Errorf("failed to get crd: %+v", obj)
		return
	}
	s.logger.Sugar().Infof("delete crd: %v", r.Name)
}

// ===================================

func (s *informerHandler) executeInformer() {

	// ------- client set
	config, err := rest.InClusterConfig()
	if err != nil {
		s.logger.Sugar().Fatalf("failed to InClusterConfig, reason=%v", err)
	}
	clientset, err := crdclientset.NewForConfig(config) // 初始化 client
	if err != nil {
		s.logger.Sugar().Fatalf("failed to NewForConfig, reason=%v", err)
	}
	s.k8sclient = clientset

	ctx, cancel := context.WithCancel(context.Background())

	if len(s.leaseName) > 0 && len(s.leaseNameSpace) > 0 && len(s.leaseId) > 0 {
		s.logger.Sugar().Infof("%v try to get lease %s/%s to run informer", s.leaseId, s.leaseNameSpace, s.leaseName)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		rlogger := s.logger.Named(fmt.Sprintf("lease %s/%s", s.leaseNameSpace, s.leaseName))
		// id := globalConfig.PodName
		getLease, lossLease, err := lease.NewLeaseElector(ctx, s.leaseNameSpace, s.leaseName, s.leaseId, rlogger)
		if err != nil {
			s.logger.Sugar().Fatalf("failed to generate lease, reason=%v ", err)
		}
		<-getLease
		s.logger.Sugar().Infof("succeed to get lease %s/%s to run informer", s.leaseNameSpace, s.leaseName)

		go func(lossLease chan struct{}) {
			<-lossLease
			cancel()
			s.logger.Sugar().Warnf("lease %s/%s is loss, informer is broken", s.leaseNameSpace, s.leaseName)
		}(lossLease)
	}

	// setup informer
	s.logger.Info("begin to setup informer")
	factory := externalversions.NewSharedInformerFactory(clientset, 0)
	// 注意，一个 factory 下  对同一种 CRD 不能 创建 多个Informer，不然会 数据竞争 问题。 而 一个 factory 下， 可对不同 CRD 产生 各种的 Informer

	t := factory.Elf().V1().Mybooks()
	s.crdlister = t.Lister()

	inform := t.Informer()

	// 在一个 Handler 逻辑中，是顺序消费所有的 crd 事件的
	// 简单说：有2个 crd add 事件，那么，先会调用 informerAddHandler 完成 事件1 后，才会 调用 informerAddHandler 处理 事件2
	inform.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.informerAddHandler,
		UpdateFunc: s.informerUpdateHandler,
		DeleteFunc: s.informerDeleteHandler,
	})

	// 一个 inform 下  如果注册 第二套 AddEventHandler，那么，对于同一个 事件，两套 handler 是 使用 独立协程 并发调用的 .
	// 这样，就能实现对同一个事件 并发调用不同的回调，好处是，他们底层是基于同一个 NewSharedInformer ， 共用一个cache，能降低api server 之间的数据同步
	inform.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.informerAddHandler,
		UpdateFunc: s.informerUpdateHandler,
		DeleteFunc: s.informerDeleteHandler,
	})

	defer s.queue.ShutDown()
	go func() {
		s.logger.Info("start worker")

		if !cache.WaitForNamedCacheSync("mybook-controller", ctx.Done(), inform.HasSynced) {
			s.logger.Error("failed to sync cache")
			cancel()
			return
		}

		for i := 0; i < workerNumber; i++ {
			go wait.UntilWithContext(ctx, s.worker, time.Second)
		}
	}()

	inform.Run(ctx.Done())

}

func (s *mybookManager) RunController(leaseName, leaseNameSpace string, leaseId string) {

	scheme, e := crd.SchemeBuilder.Build()
	if e != nil {
		s.logger.Sugar().Fatalf("failed to get crd scheme: %+v", e)
	}
	/*
		Events:
		  Type    Reason     Age   From    Message
		  ----    ------     ----  ----    -------
		  Normal  newMybook  13s   mybook  crd event, new mybook test
	*/
	p := k8s.NewEventRecord(scheme, "mybook", s.logger)

	// -----------
	t := &informerHandler{
		logger:         s.logger,
		leaseName:      leaseName,
		leaseNameSpace: leaseNameSpace,
		leaseId:        leaseId,
		eventRecord:    p,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mybook"),
	}
	s.informer = t

	go func() {
		for {
			t.executeInformer()
			time.Sleep(time.Duration(5) * time.Second)
		}
	}()
}
