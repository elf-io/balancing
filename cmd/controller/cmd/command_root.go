// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

var BinName = filepath.Base(os.Args[0])
var rootLogger *zap.Logger

// rootCmd represents the base command.
var rootCmd = &cobra.Command{
	Use:   BinName,
	Short: "short description",
	Run: func(cmd *cobra.Command, args []string) {
		DaemonMain()
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	if _, err := maxprocs.Set(
		maxprocs.Logger(func(s string, i ...interface{}) {
			rootLogger.Sugar().Infof(s, i...)
		}),
	); err != nil {
		rootLogger.Sugar().Warn("failed to set GOMAXPROCS")
	}

	if err := rootCmd.Execute(); err != nil {
		rootLogger.Fatal(err.Error())
	}
}
