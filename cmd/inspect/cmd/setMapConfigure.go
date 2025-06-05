// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/spf13/cobra"
	"os"
)

var CmdSetMapConfigure = &cobra.Command{
	Use:   "configure",
	Short: "set the ebpf map of configure ",
}

var CmdSetMapConfigureDebugLevel = &cobra.Command{
	Use:   "debugLevel",
	Short: "set debug level: verbose/info/error ",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("supported argument: verbose/info/error ")
			os.Exit(1)
		}

		var key, value uint32
		key = ebpf.MapConfigureKeyIndexDebugLevel
		switch args[0] {
		case "verbose":
			value = ebpf.MapConfigureValueDebugLevelVerbose
		case "info":
			value = ebpf.MapConfigureValueDebugLevelInfo
		case "error":
			value = ebpf.MapConfigureValueDebugLevelError
		default:
			fmt.Printf("supported argument: verbose/info/error ")
			os.Exit(1)
		}

		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()
		if e := bpf.UpdateMapConfigure(key, value); e != nil {
			fmt.Printf("failed to set ebpf Map: %v\n", e)
			os.Exit(3)
		}
		fmt.Printf("succeeded to set ebpf Map configure: %s\n", ebpf.MapConfigureStr(key, value))
	},
}

var CmdSetMapConfigureIpv4 = &cobra.Command{
	Use:   "ipv4Enabled",
	Short: "set ipv4 enabled: on/off ",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("supported argument: on/off ")
			os.Exit(1)
		}

		var key, value uint32
		key = ebpf.MapConfigureKeyIndexIpv4Enabled
		switch args[0] {
		case "on":
			value = ebpf.MapConfigureValueEnabled
		case "off":
			value = ebpf.MapConfigureValueDisabled
		default:
			fmt.Printf("supported argument: on/off ")
			os.Exit(1)
		}

		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()
		if e := bpf.UpdateMapConfigure(key, value); e != nil {
			fmt.Printf("failed to set ebpf Map: %v\n", e)
			os.Exit(3)
		}
		fmt.Printf("succeeded to set ebpf Map configure: %s\n", ebpf.MapConfigureStr(key, value))
	},
}

var CmdSetMapConfigureIpv6 = &cobra.Command{
	Use:   "ipv6Enabled",
	Short: "set ipv6 enabled: on/off ",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("supported argument: on/off ")
			os.Exit(1)
		}

		var key, value uint32
		key = ebpf.MapConfigureKeyIndexIpv6Enabled
		switch args[0] {
		case "on":
			value = ebpf.MapConfigureValueEnabled
		case "off":
			value = ebpf.MapConfigureValueDisabled
		default:
			fmt.Printf("supported argument: on/off ")
			os.Exit(1)
		}

		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()
		if e := bpf.UpdateMapConfigure(key, value); e != nil {
			fmt.Printf("failed to set ebpf Map: %v\n", e)
			os.Exit(3)
		}
		fmt.Printf("succeeded to set ebpf Map configure: %s\n", ebpf.MapConfigureStr(key, value))
	},
}

var CmdSetMapConfigureRedirectQosLimit = &cobra.Command{
	Use:   "redirectQosLimit",
	Short: "set redirect QoS limit (requests per second)",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("please provide a numeric value for redirectQosLimit")
			os.Exit(1)
		}

		var key uint32
		var value uint32
		key = ebpf.MapConfigureKeyIndexRedirectQoSLimit

		// Convert string argument to uint32
		_, err := fmt.Sscanf(args[0], "%d", &value)
		if err != nil {
			fmt.Printf("Invalid value: %s. Please provide a valid number\n", args[0])
			os.Exit(1)
		}

		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		if e := bpf.UpdateMapConfigure(key, value); e != nil {
			fmt.Printf("failed to set ebpf Map: %v\n", e)
			os.Exit(3)
		}

		fmt.Printf("succeeded to set ebpf Map configure: %s\n", ebpf.MapConfigureStr(key, value))
	},
}

func init() {
	CmdSetMapConfigure.AddCommand(CmdSetMapConfigureIpv4)
	CmdSetMapConfigure.AddCommand(CmdSetMapConfigureIpv6)
	CmdSetMapConfigure.AddCommand(CmdSetMapConfigureDebugLevel)
	CmdSetMapConfigure.AddCommand(CmdSetMapConfigureRedirectQosLimit)
	CmdSetMap.AddCommand(CmdSetMapConfigure)
}
