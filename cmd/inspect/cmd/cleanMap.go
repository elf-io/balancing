// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
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
