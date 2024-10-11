package cmd

import (
	"fmt"
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

// for CRD
func SetupController() {

	// ctrl.SetLogger(logr.New(controllerruntimelog.NullLogSink{}))
	ctrl.SetLogger(controllerzap.New())

	// controller for CRD
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// Readiness probe endpoint name, defaults to "readyz"
		// Liveness probe endpoint name, defaults to "healthz"
		HealthProbeBindAddress:  fmt.Sprintf("0.0.0.0:%d", types.ControllerConfig.HttpPort),
		LeaderElection:          true,
		LeaderElectionID:        "balacning-leader",
		LeaderElectionNamespace: types.ControllerConfig.PodNamespace,
	})
	if err != nil {
		rootLogger.Sugar().Fatalf("unable to NewManager: %v", err)
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

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		rootLogger.Sugar().Fatalf("problem running manager: %v", err)
	}

}
