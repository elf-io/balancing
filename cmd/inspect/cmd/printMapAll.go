package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/elf-io/balancing/pkg/ebpf"
	"os"
)

var CmdPrintMapAll = &cobra.Command{
	Use:   "all",
	Short: "print all data of the ebpf map ",
	Args:  cobra.RangeArgs(0, 0),
	Run: func(cmd *cobra.Command, args []string) {
		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		fmt.Printf("\n")
		fmt.Printf("print all data of the ebpf map:\n")
		bpf.PrintMapAffinity()
		bpf.PrintMapNatRecord()
		bpf.PrintMapService()
		bpf.PrintMapBackend()
		bpf.PrintMapNodeIp()
		bpf.PrintMapNodeEntryIp()
		fmt.Printf("\n")
	},
}

func init() {
	CmdPrintMap.AddCommand(CmdPrintMapAll)
}
