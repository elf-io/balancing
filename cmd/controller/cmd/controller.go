package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(balancingv1beta1.AddToScheme(scheme))
}

func HealthCheckHandler(req *http.Request) error {
	if finishSetUp {
		return nil
	}
	return fmt.Errorf("setting up")
}

var finishSetUp = false

// for CRD
func SetupController() {

	// ctrl.SetLogger(logr.New(controllerruntimelog.NullLogSink{}))
	ctrl.SetLogger(controllerzap.New())

	// controller for CRD
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// Readiness probe endpoint name, defaults to "readyz"
		// Liveness probe endpoint name, defaults to "healthz"
		HealthProbeBindAddress:  fmt.Sprintf(":%d", types.ControllerConfig.HttpPort),
		LeaderElection:          true,
		LeaderElectionID:        "balacning-leader",
		LeaderElectionNamespace: types.ControllerConfig.PodNamespace,
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

	t := &webhookBalacning{}
	err = ctrl.NewWebhookManagedBy(mgr).
		For(&balancingv1beta1.BalancingPolicy{}).
		WithDefaulter(t).
		WithValidator(t).
		Complete()
	if err != nil {
		rootLogger.Sugar().Fatalf("unable to NewWebhookManagedBy for BalancingPolicy : %v", err)
	}

	m := &webhookRedirect{}
	err = ctrl.NewWebhookManagedBy(mgr).
		For(&balancingv1beta1.LocalRedirectPolicy{}).
		WithDefaulter(m).
		WithValidator(m).
		Complete()
	if err != nil {
		rootLogger.Sugar().Fatalf("unable to NewWebhookManagedBy for LocalRedirectPolicy : %v", err)
	}

	go func() {
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			rootLogger.Sugar().Fatalf("problem running manager: %v", err)
		}
	}()
	finishSetUp = true

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	for sig := range sigCh {
		rootLogger.Sugar().Warnf("Received singal %+v ", sig)
		os.Exit(1)
	}
}
