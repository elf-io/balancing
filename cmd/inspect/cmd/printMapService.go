package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/elf-io/balancing/pkg/ebpf"
	"os"
)

var CmdPrintMapService = &cobra.Command{
	Use:   "service",
	Short: "print the ebpf map of service ",

	Args: cobra.RangeArgs(0, 0),

	Run: func(cmd *cobra.Command, args []string) {

		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		fmt.Printf("\n")
		fmt.Printf("print the ebpf map of service:\n")
		bpf.PrintMapService()
		fmt.Printf("\n")
	},
}

func init() {
	CmdPrintMap.AddCommand(CmdPrintMapService)
}
