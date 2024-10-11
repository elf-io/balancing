// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/elf-io/balancing/pkg/ebpfEvent"
	"github.com/elf-io/balancing/pkg/ebpfWriter"
	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/nodeId"
	"github.com/elf-io/balancing/pkg/podId"
	"github.com/elf-io/balancing/pkg/podLabel"
	"github.com/elf-io/balancing/pkg/types"
	"github.com/elf-io/balancing/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"time"
)

/*
type reconciler struct {
	// client can be used to retrieve objects from the APIServer.
	client client.Client
	log    *zap.Logger
}

func (r *reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	t := reconcile.Result{}

	r.log.Sugar().Infof("Reconcile: %v", req)

	return t, nil
}

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func SetupController() {
	logger := rootLogger.Named("controller")

	config := ctrl.GetConfigOrDie()
	config.Burst = 100
	config.QPS = 50
	mgr, err := ctrl.NewManager(config, manager.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
	})
	if err != nil {
		logger.Sugar().Fatalf("unable to set up controller: %v ", err)
	}

	ctrl.SetLogger(k8szap.New())

	r := reconciler{
		client: mgr.GetClient(),
		log:    logger,
	}
	// Setup a new controller to reconcile ReplicaSets
	logger.Sugar().Info("Setting up controller")
	c, err := controller.New("agent", mgr, controller.Options{
		Reconciler: &r,
	})
	if err != nil {
		logger.Sugar().Fatalf("unable to set up individual controller: %v", err)
	}

	// Watch ReplicaSets and enqueue ReplicaSet object key
	if err := c.Watch(source.Kind(mgr.GetCache(), &corev1.Service{}, &handler.TypedEnqueueRequestForObject[*corev1.Service]{})); err != nil {
		logger.Sugar().Fatalf("unable to watch service: %v", err)
	}
	if err := c.Watch(source.Kind(mgr.GetCache(), &discovery.EndpointSlice{}, &handler.TypedEnqueueRequestForObject[*discovery.EndpointSlice]{})); err != nil {
		logger.Sugar().Fatalf("unable to watch EndpointSlice: %v", err)
	}

	logger.Sugar().Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Sugar().Fatalf("unable to run manager: %v", err)
	}

}
*/
// ------------------------------

var (
	InformerListInvterval = time.Second * 60
)

func RunReconciles() {

	rootLogger.Sugar().Debugf("RunReconciles")

	// get clientset
	c, e1 := utils.AutoK8sConfig()
	if e1 != nil {
		rootLogger.Sugar().Fatalf("failed to find client-go config, make sure it is in a pod or ~/.kube/config exists: %v", e1)
	}
	Client, e2 := kubernetes.NewForConfig(c)
	if e2 != nil {
		rootLogger.Sugar().Fatalf("failed to NewForConfig: %v", e2)
	}

	// before informer and ebpf, build nodeId database
	nodeId.InitNodeIdManager(Client, rootLogger.Named("nodeId"))

	// before informer and ebpf, build pod ip database of local node
	podId.InitPodIdManager(Client, rootLogger.Named("podId"), types.AgentConfig.LocalNodeName)

	podLabel.InitPodLabelManager(rootLogger.Named("podLabel"))

	// setup ebpf and load
	bpfManager := ebpf.NewEbpfProgramMananger(rootLogger.Named("ebpf"))
	if err := bpfManager.LoadProgramp(); err != nil {
		rootLogger.Sugar().Fatalf("failed to Load ebpf Programp: %v \n", err)
	}
	defer bpfManager.UnloadProgramp()
	rootLogger.Sugar().Infof("succeeded to Load ebpf Programp \n")
	// setup ebpf writer
	writer := ebpfWriter.NewEbpfWriter(Client, bpfManager, InformerListInvterval, rootLogger.Named("ebpfWriter"))
	// before informer, clean all map data to keep all data up to date
	writer.CleanEbpfMapData()

	// setup informer
	stopWatchCh := make(chan struct{})
	NewPodInformer(Client, stopWatchCh, types.AgentConfig.LocalNodeName, writer)
	NewNodeInformer(Client, stopWatchCh, writer)

	NewServiceInformer(Client, stopWatchCh, writer)
	NewEndpointSliceInformer(Client, stopWatchCh, writer)

	// crd reconcile
	SetupController(writer)

	//
	ebpfEvent := ebpfEvent.NewEbpfEvent(rootLogger.Named("ebpfEvent"), bpfManager)
	ebpfEvent.WatchEbpfEvent(stopWatchCh)

	rootLogger.Info("finish all setup ")
	time.Sleep(time.Hour)

}

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(balancingv1beta1.AddToScheme(scheme))
}

// for CRD
func SetupController(writer ebpfWriter.EbpfWriter) {

	// ctrl.SetLogger(logr.New(controllerruntimelog.NullLogSink{}))
	ctrl.SetLogger(controllerzap.New())

	// controller for CRD
	rootLogger.Info("setup crd controller ")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// todo: metric server
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		// todo: HealthProbe
		HealthProbeBindAddress: fmt.Sprintf("0.0.0.0:%d", types.AgentConfig.HttpPort),
	})
	if err != nil {
		rootLogger.Sugar().Fatalf("unable to NewManager: %v", err)
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&balancingv1beta1.BalancingPolicy{}).
		Complete(&ReconcilerBalancing{
			client: mgr.GetClient(),
			l:      rootLogger.Named("BalancingPolicyReconciler"),
			writer: writer,
		})
	if err != nil {
		rootLogger.Sugar().Fatalf("unable to NewControllerManagedBy for BalancingPolicy: %v", err)
	}
	rootLogger.Info("setup controller for BalancingPolicy")

	err = ctrl.NewControllerManagedBy(mgr).
		For(&balancingv1beta1.LocalRedirectPolicy{}).
		Complete(&ReconcilerRedirect{
			client: mgr.GetClient(),
			l:      rootLogger.Named("LocalRedirectPolicyReconciler"),
			writer: writer,
		})
	if err != nil {
		rootLogger.Sugar().Fatalf("unable to NewControllerManagedBy for LocalRedirectPolicy : %v", err)
	}
	rootLogger.Info("setup controller for LocalRedirectPolicy")

	go func() {
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			rootLogger.Sugar().Fatalf("problem running crd controller: %v", err)
		}
		rootLogger.Warn("crd controller exits")
	}()

	waitForCacheSync := mgr.GetCache().WaitForCacheSync(context.Background())
	if !waitForCacheSync {
		rootLogger.Fatal("failed to wait for syncing controller-runtime cache")
	}

}
