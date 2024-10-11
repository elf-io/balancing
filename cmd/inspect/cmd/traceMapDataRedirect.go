package cmd

import (
	"context"
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/elf-io/balancing/pkg/ebpfWriter"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"

	crdclientset "github.com/elf-io/balancing/pkg/k8s/client/clientset/versioned/typed/balancing.elf.io/v1beta1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var CmdTraceMapByRedirect = &cobra.Command{
	Use:   "localRedirect  policyName",
	Short: "get all the ebpf map data relevant to the localredirect policy",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		policyName := args[0]
		fmt.Printf("get all the ebpf map data relevant to the localredirect policy %s \n", policyName)
		fmt.Printf("\n")

		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		// Load kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			log.Fatalf("Failed to load kubeconfig: %v", err)
		}

		crdClient, err := crdclientset.NewForConfig(config)
		if err != nil {
			log.Fatalf("Failed to create Kubernetes client: %v", err)
		}
		// Query the specified service
		policy, err := crdClient.LocalRedirectPolicies().Get(context.TODO(), policyName, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Failed to get policy %s: %v", policyName, err)
		}

		var svcV4Id uint32
		if policy.Spec.RedirectFrontend.ServiceMatcher != nil {
			k8sClient, err := kubernetes.NewForConfig(config)
			if err != nil {
				log.Fatalf("Failed to create Kubernetes client: %v", err)
			}

			t := policy.Spec.RedirectFrontend.ServiceMatcher
			// Query the specified service
			service, err := k8sClient.CoreV1().Services(t.Namespace).Get(context.TODO(), t.ServiceName, metav1.GetOptions{})
			if err != nil {
				log.Fatalf("Failed to get service %s/%s: %v", t.Namespace, t.ServiceName, err)
			}
			// for ipv4 data
			svcV4Id = ebpf.GenerateSvcV4Id(service)
			// todo: for ipv6 id
		} else {
			if t, e := ebpfWriter.FakeServiceByAddressMatcher(policy); e != nil {
				log.Fatalf("Failed to fake service for RedirectPolicy %v", e)
			} else {
				// for ipv4 data
				svcV4Id = ebpf.GenerateSvcV4Id(t)
				// todo: for ipv6 id
			}
		}

		bpf.PrintMapService(&ebpf.NAT_TYPE_REDIRECT, &svcV4Id)
		fmt.Printf("\n")

		bpf.PrintMapBackend(&ebpf.NAT_TYPE_REDIRECT, &svcV4Id)
		fmt.Printf("\n")

		// todo: ipv6

	},
}

func init() {
	CmdTraceMap.AddCommand(CmdTraceMapByRedirect)
}
