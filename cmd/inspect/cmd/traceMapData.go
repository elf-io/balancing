// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"github.com/spf13/cobra"
)

var CmdTraceMap = &cobra.Command{
	Use:   "traceMapData",
	Short: "trace ebpf map data by the context",
}

func init() {
	RootCmd.AddCommand(CmdTraceMap)
}
