package cmd

import (
	"context"
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpf"
	"log"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var CmdTraceMapByService = &cobra.Command{
	Use:   "service Namespace ServiceName",
	Short: "get all the ebpf map data relevant to the service",
	Args:  cobra.RangeArgs(2, 2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		serviceName := args[1]
		fmt.Printf("trace the service data of ebpf map for the service %s/%s \n", namespace, serviceName)
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

		// Create a new Kubernetes client
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Fatalf("Failed to create Kubernetes client: %v", err)
		}

		// Query the specified service
		service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Failed to get service %s/%s: %v", namespace, serviceName, err)
		}

		// for ipv4 data
		svcV4Id := ebpf.GenerateSvcV4Id(service)
		if svcV4Id == 0 {
			fmt.Printf("the service %s/%s does not have ipv4 data\n", namespace, serviceName)
		} else {
			bpf.PrintMapService(&ebpf.NAT_TYPE_SERVICE, &svcV4Id)
			fmt.Printf("\n")

			bpf.PrintMapBackend(&ebpf.NAT_TYPE_SERVICE, &svcV4Id)
			fmt.Printf("\n")
		}

		// todo: ipv6

	},
}

func init() {
	CmdTraceMap.AddCommand(CmdTraceMapByService)
}
