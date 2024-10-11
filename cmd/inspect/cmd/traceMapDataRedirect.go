package cmd

import (
	"context"
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/elf-io/balancing/pkg/types"
	"github.com/elf-io/balancing/pkg/utils"
	"log"
	"os"

	crdclientset "github.com/elf-io/balancing/pkg/k8s/client/clientset/versioned/typed/balancing.elf.io/v1beta1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		config, err := utils.AutoCrdConfig()
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

		// get the serviceId
		idStr, ok := policy.Annotations[types.AnnotationServiceID]
		if !ok {
			log.Fatalf("Failed to get serviceId annotation from policy %s", policyName)
		}
		svcV4Id, e := utils.StringToUint32(idStr)
		if e != nil {
			log.Fatalf("Failed to generate serviceId from policy %s: %s ", policyName, idStr)
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
