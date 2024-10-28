// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/spf13/cobra"
	"os"
)

var CmdPrintMapAll = &cobra.Command{
	Use:   "all",
	Short: "print all data of the ebpf map ",
	Args:  cobra.RangeArgs(0, 0),
	Run: func(cmd *cobra.Command, args []string) {
		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		fmt.Printf("\n")
		fmt.Printf("print all data of the ebpf map:\n")
		bpf.PrintMapAffinity()
		bpf.PrintMapNatRecord()
		bpf.PrintMapService(nil, nil)
		bpf.PrintMapBackend(nil, nil)
		bpf.PrintMapNodeIp()
		bpf.PrintMapNodeProxyIp()
		bpf.PrintMapConfigure()
		fmt.Printf("\n")
	},
}

func init() {
	CmdPrintMap.AddCommand(CmdPrintMapAll)
}
