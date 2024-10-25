// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpfWriter"
	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/lock"
	"github.com/elf-io/balancing/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
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
func SetupController(clientConfig *rest.Config, writer ebpfWriter.EbpfWriter) {

	// ctrl.SetLogger(logr.New(controllerruntimelog.NullLogSink{}))
	ctrl.SetLogger(controllerzap.New())

	// controller for CRD
	rootLogger.Info("setup crd controller ")
	// c:=ctrl.GetConfigOrDie()
	c := clientConfig
	mgr, err := ctrl.NewManager(c, ctrl.Options{
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
