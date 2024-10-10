package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var CmdTraceMapByService = &cobra.Command{
	Use:   "service [-n Namespace] serviceName",
	Short: "clean the ebpf map of affinity ",
	Args:  cobra.RangeArgs(2, 2),
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("args: %+v\n", args)

	},
}

func init() {
	CmdTraceMap.AddCommand(CmdTraceMapByService)
}
