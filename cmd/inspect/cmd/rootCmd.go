// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var BinName = filepath.Base(os.Args[0])

// var rootLogger *zap.Logger

var RootCmd = &cobra.Command{
	Use:   BinName,
	Short: "cli for debugging",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		panic(err.Error())
	}
}
