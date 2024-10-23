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
	"github.com/elf-io/balancing/pkg/lock"
	"github.com/elf-io/balancing/pkg/nodeId"
	"github.com/elf-io/balancing/pkg/podId"
	"github.com/elf-io/balancing/pkg/podLabel"
	"github.com/elf-io/balancing/pkg/types"
	"github.com/elf-io/balancing/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"os"
	"os/signal"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"syscall"
	"time"
)

var (
	InformerListInvterval = time.Second * 60
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(balancingv1beta1.AddToScheme(scheme))
}

var finishSetUp = false
var finishlock = &lock.RWMutex{}

func HealthCheckHandler(req *http.Request) error {
	finishlock.RLock()
	defer finishlock.RUnlock()
	if finishSetUp {
		return nil
	}

	rootLogger.Sugar().Warnf("health is not ready")
	return fmt.Errorf("setting up")
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
		HealthProbeBindAddress: fmt.Sprintf(":%d", types.AgentConfig.HttpPort),
	})
	if err != nil {
		rootLogger.Sugar().Fatalf("unable to NewManager: %v", err)
	}

	// for liveness check, with url "/healthz"
	if err := mgr.AddHealthzCheck("healthz", HealthCheckHandler); err != nil {
		rootLogger.Sugar().Fatalf("unable to set up liveness check: %v", err)
	}

	// for readiness check, with url "/readyz"
	if err := mgr.AddReadyzCheck("readyz", HealthCheckHandler); err != nil {
		rootLogger.Sugar().Fatalf("unable to set up readiness check: %v", err)
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

func RunReconciles() {

	rootLogger.Sugar().Debugf("RunReconciles")

	// get clientset
	c, e1 := utils.AutoK8sConfig(types.AgentConfig.KubeconfigPath)
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
	rootLogger.Sugar().Infof("succeeded to Load ebpf Programp \n")
	// setup ebpf writer
	writer := ebpfWriter.NewEbpfWriter(Client, bpfManager, InformerListInvterval, rootLogger.Named("ebpfWriter"))
	// before informer, clean all map data to keep all data up to date
	writer.CleanEbpfMapData()

	// setup informer
	stopWatchCh := make(chan struct{})
	NewPodInformer(Client, stopWatchCh, writer)
	NewNodeInformer(Client, stopWatchCh, writer)

	NewServiceInformer(Client, stopWatchCh, writer)
	NewEndpointSliceInformer(Client, stopWatchCh, writer)

	// crd reconcile
	SetupController(writer)

	//
	ebpfEvent := ebpfEvent.NewEbpfEvent(rootLogger.Named("ebpfEvent"), bpfManager, writer)
	ebpfEvent.WatchEbpfEvent(stopWatchCh)

	rootLogger.Info("finish all setup ")

	finishlock.Lock()
	finishSetUp = true
	finishlock.Unlock()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	for sig := range sigCh {
		rootLogger.Sugar().Warnf("Received singal %+v ", sig)
		bpfManager.UnloadProgramp()
		os.Exit(1)
	}

}
