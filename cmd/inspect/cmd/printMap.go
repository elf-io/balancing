// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"github.com/spf13/cobra"
)

var CmdPrintMap = &cobra.Command{
	Use:   "showMapData",
	Short: "show the data of ebpf map",
}

func init() {
	RootCmd.AddCommand(CmdPrintMap)
}
