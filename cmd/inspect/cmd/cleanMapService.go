package cmd

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/spf13/cobra"
	"os"
)

var CmdCleanMapService = &cobra.Command{
	Use:   "service",
	Short: "clean the ebpf map of service ",
	Args:  cobra.RangeArgs(0, 0),
	Run: func(cmd *cobra.Command, args []string) {
		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		fmt.Printf("\n")
		fmt.Printf("clean the ebpf map of service:\n")
		if c, e := bpf.CleanMapService(); e != nil {
			fmt.Printf("    failed to clean: %+v\n", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items\n", c)
		}
		fmt.Printf("\n")
	},
}

func init() {
	CmdCleanMap.AddCommand(CmdCleanMapService)
}
