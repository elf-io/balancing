package cmd

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/utils"
	"net/http"
	"path"
	"time"

	balancingv1beta1 "github.com/elf-io/balancing/pkg/k8s/apis/balancing.elf.io/v1beta1"
	"github.com/elf-io/balancing/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	runtimeWebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	InformerListInvterval = time.Second * 60 * 5
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(balancingv1beta1.AddToScheme(scheme))
}

func HealthCheckHandler(req *http.Request) error {
	if finishSetUp {
		return nil
	}
	rootLogger.Sugar().Warnf("health is not ready")
	return fmt.Errorf("setting up")
}

var finishSetUp = false

// for CRD
func SetupController() {

	// get clientset
	apiServerHostAddress := ""
	if len(types.ControllerConfig.Configmap.ApiServerHost) > 0 && len(types.ControllerConfig.Configmap.ApiServerPort) > 0 {
		apiServerHostAddress = fmt.Sprintf("%s:%s", types.ControllerConfig.Configmap.ApiServerHost, types.ControllerConfig.Configmap.ApiServerPort)
		rootLogger.Sugar().Infof("in cluster: replace the address of api Server to %s", apiServerHostAddress)
	}
	clientConfig, e1 := utils.AutoK8sConfig("", apiServerHostAddress)
	if e1 != nil {
		rootLogger.Sugar().Fatalf("failed to find client-go config, make sure it is in a pod or ~/.kube/config exists: %v", e1)
	}
	rootLogger.Sugar().Debugf("clientConfig: %+v", clientConfig)

	// ctrl.SetLogger(logr.New(controllerruntimelog.NullLogSink{}))
	ctrl.SetLogger(controllerzap.New())

	// controller for CRD
	// c:=ctrl.GetConfigOrDie()
	c := clientConfig
	mgr, err := ctrl.NewManager(c, ctrl.Options{
		Scheme: scheme,
		// Readiness probe endpoint name, defaults to "readyz"
		// Liveness probe endpoint name, defaults to "healthz"
		HealthProbeBindAddress:  fmt.Sprintf(":%d", types.ControllerConfig.HttpPort),
		LeaderElection:          true,
		LeaderElectionID:        "balacning-leader",
		LeaderElectionNamespace: types.ControllerConfig.PodNamespace,
		WebhookServer: runtimeWebhook.NewServer(runtimeWebhook.Options{
			Port:     int(types.ControllerConfig.WebhookPort),
			CertDir:  path.Dir(types.ControllerConfig.TlsCaCertPath),
			CertName: path.Base(types.ControllerConfig.TlsServerCertPath),
			KeyName:  path.Base(types.ControllerConfig.TlsServerKeyPath),
			// ClientCAName is the CA certificate name which server used to verify remote(client)'s certificate.
			// Defaults to "", which means server does not verify client's certificate.
			// ClientCAName:  path.Base(types.ControllerConfig.TlsCaCertPath),
			ClientCAName: "",
		}),
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

	t := &webhookBalacning{
		l: rootLogger.Named("balancingWebhook"),
	}
	err = ctrl.NewWebhookManagedBy(mgr).
		For(&balancingv1beta1.BalancingPolicy{}).
		WithDefaulter(t).
		WithValidator(t).
		Complete()
	if err != nil {
		rootLogger.Sugar().Fatalf("unable to NewWebhookManagedBy for BalancingPolicy : %v", err)
	}

	m := &webhookRedirect{
		l: rootLogger.Named("localRedirectWebhook"),
	}
	err = ctrl.NewWebhookManagedBy(mgr).
		For(&balancingv1beta1.LocalRedirectPolicy{}).
		WithDefaulter(m).
		WithValidator(m).
		Complete()
	if err != nil {
		rootLogger.Sugar().Fatalf("unable to NewWebhookManagedBy for LocalRedirectPolicy : %v", err)
	}

	go func() {
		rootLogger.Sugar().Infof("begin to run controller")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			rootLogger.Sugar().Fatalf("problem running manager: %v", err)
		}
	}()
	finishSetUp = true

}
