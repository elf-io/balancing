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
		if err := bpf.PrintMapAffinity(); err != nil {
			fmt.Println("Error:", err)
		}
		if err := bpf.PrintMapNatRecord(); err != nil {
			fmt.Println("Error:", err)
		}
		if err := bpf.PrintMapService(nil, nil); err != nil {
			fmt.Println("Error:", err)
		}
		if err := bpf.PrintMapBackend(nil, nil); err != nil {
			fmt.Println("Error:", err)
		}
		if err := bpf.PrintMapNodeIp(); err != nil {
			fmt.Println("Error:", err)
		}
		if err := bpf.PrintMapNodeProxyIp(); err != nil {
			fmt.Println("Error:", err)
		}
		if err := bpf.PrintMapConfigure(); err != nil {
			fmt.Println("Error:", err)
		}
		fmt.Printf("\n")
	},
}

func init() {
	CmdPrintMap.AddCommand(CmdPrintMapAll)
}
