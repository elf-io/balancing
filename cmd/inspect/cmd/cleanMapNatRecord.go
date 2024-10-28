// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/elf-io/balancing/pkg/ebpf"
	"os"
)

var CmdCleanMapNat = &cobra.Command{
	Use:   "natRecord",
	Short: "clean the ebpf map of natRecord ",
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
		fmt.Printf("clean the ebpf map of natRecord:
")
		if c, e := bpf.CleanMapNatRecord(); e != nil {
			fmt.Printf("    failed to clean: %+v
", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items
", c)
		}
		fmt.Printf("
")
	},
}

func init() {
	CmdCleanMap.AddCommand(CmdCleanMapNat)
}
