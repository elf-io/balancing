package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/elf-io/elf/pkg/ebpf"
	"os"
)

var CmdPrintMapAffinity = &cobra.Command{
	Use:   "affinity",
	Short: "print the ebpf map of affinity ",
	Args:  cobra.RangeArgs(0, 0),
	Run: func(cmd *cobra.Command, args []string) {
		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		fmt.Printf("\n")
		fmt.Printf("print the ebpf map of affinity:\n")
		bpf.PrintMapAffinity()
		fmt.Printf("\n")
	},
}

func init() {
	CmdPrintMap.AddCommand(CmdPrintMapAffinity)
}
