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
			fmt.Printf("failed to load ebpf Map: %v
", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		fmt.Printf("
")
		fmt.Printf("clean all ebpf map:
")
		if c, e := bpf.CleanMapService(); e != nil {
			fmt.Printf("    failed to clean service map: %+v
", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in service map 
", c)
		}
		if c, e := bpf.CleanMapBackend(); e != nil {
			fmt.Printf("    failed to clean backend map: %+v
", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in backend map
", c)
		}
		if c, e := bpf.CleanMapNodeIp(); e != nil {
			fmt.Printf("    failed to clean node nodeIp: %+v
", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in nodeIp map
", c)
		}
		if c, e := bpf.CleanMapNodeProxyIp(); e != nil {
			fmt.Printf("    failed to clean nodeProxyIp map: %+v
", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in nodeProxyIp map
", c)
		}
		if c, e := bpf.CleanMapNatRecord(); e != nil {
			fmt.Printf("    failed to clean natRecord map: %+v
", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in natRecord map
", c)
		}
		if c, e := bpf.CleanMapAffinity(); e != nil {
			fmt.Printf("    failed to clean affinity map: %+v
", e)
			os.Exit(3)
		} else {
			fmt.Printf("    succeeded to clean %d items in affinity map
", c)
		}

		fmt.Printf("
")
	},
}

func init() {
	CmdCleanMap.AddCommand(CmdCleanMapAll)
}
