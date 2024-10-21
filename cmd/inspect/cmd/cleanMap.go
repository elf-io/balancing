package cmd

import (
	"github.com/spf13/cobra"
)

var CmdCleanMap = &cobra.Command{
	Use:   "cleanMapData",
	Short: "clean the data of ebpf map",
}

func init() {
	RootCmd.AddCommand(CmdCleanMap)
}
