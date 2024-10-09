package cmd

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/spf13/cobra"
	"os"
)

var CmdPrintMapNodeProxyIp = &cobra.Command{
	Use:   "nodeProxyIp",
	Short: "print the ebpf map of nodeProxyIp ",
	Args:  cobra.RangeArgs(0, 0),
	Run: func(cmd *cobra.Command, args []string) {
		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		fmt.Printf("\n")
		fmt.Printf("print the ebpf map of nodeProxyIp:\n")
		bpf.PrintMapNodeProxyIp()
		fmt.Printf("\n")
	},
}

func init() {
	CmdPrintMap.AddCommand(CmdPrintMapNodeProxyIp)
}
