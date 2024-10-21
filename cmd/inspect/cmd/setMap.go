package cmd

import (
	"github.com/spf13/cobra"
)

var CmdSetMap = &cobra.Command{
	Use:   "setMapData",
	Short: "set the data of ebpf map",
}

func init() {
	RootCmd.AddCommand(CmdSetMap)
}
