// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/elf-io/balancing/pkg/ebpf"
	"os"
)

var CmdPrintMapAffinity = &cobra.Command{
	Use:   "affinity",
	Short: "print the ebpf map of affinity ",
	Args:  cobra.RangeArgs(0, 0),
	Run: func(cmd *cobra.Command, args []string) {
		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v
", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		fmt.Printf("
")
		fmt.Printf("print the ebpf map of affinity:
")
		bpf.PrintMapAffinity()
		fmt.Printf("
")
	},
}

func init() {
	CmdPrintMap.AddCommand(CmdPrintMapAffinity)
}
