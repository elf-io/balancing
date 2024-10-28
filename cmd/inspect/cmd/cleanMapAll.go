// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/spf13/cobra"
	"os"
)

var CmdCleanMapAll = &cobra.Command{
	Use:   "all",
	Short: "clean all ebpf map ",
	Args:  cobra.RangeArgs(0, 0),
	Run: func(cmd *cobra.Command, args []string) {
		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		fmt.Printf("\n")
		fmt.Printf("clean all ebpf map:\n")
		if c, e := bpf.CleanMapService(); e != nil {
			fmt.Printf("    failed to clean service map: %+v\n", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in service map \n", c)
		}
		if c, e := bpf.CleanMapBackend(); e != nil {
			fmt.Printf("    failed to clean backend map: %+v\n", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in backend map\n", c)
		}
		if c, e := bpf.CleanMapNodeIp(); e != nil {
			fmt.Printf("    failed to clean node nodeIp: %+v\n", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in nodeIp map\n", c)
		}
		if c, e := bpf.CleanMapNodeProxyIp(); e != nil {
			fmt.Printf("    failed to clean nodeProxyIp map: %+v\n", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in nodeProxyIp map\n", c)
		}
		if c, e := bpf.CleanMapNatRecord(); e != nil {
			fmt.Printf("    failed to clean natRecord map: %+v\n", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in natRecord map\n", c)
		}
		if c, e := bpf.CleanMapAffinity(); e != nil {
			fmt.Printf("    failed to clean affinity map: %+v\n", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in affinity map\n", c)
		}

		fmt.Printf("\n")
	},
}

func init() {
	CmdCleanMap.AddCommand(CmdCleanMapAll)
}
