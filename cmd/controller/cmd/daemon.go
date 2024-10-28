// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/elf-io/balancing/pkg/debug"
	"github.com/elf-io/balancing/pkg/types"
	"os"
	"os/signal"
	runtimedebug "runtime/debug"
	"syscall"
)

func SetupUtility() {

	// run gops
	d := debug.New(rootLogger)
	if types.ControllerConfig.GopsPort != 0 {
		d.RunGops(int(types.ControllerConfig.GopsPort))
	}

	if types.ControllerConfig.PyroscopeServerAddress != "" {
		d.RunPyroscope(types.ControllerConfig.PyroscopeServerAddress, types.ControllerConfig.PodName)
	}
}

func DaemonMain() {

	defer func() {
		if e := recover(); e != nil {
			rootLogger.Sugar().Errorf("Panic details: %v", e)
			runtimedebug.PrintStack()
			os.Exit(1)
		}
	}()

	rootLogger.Sugar().Infof("config: %+v", types.ControllerConfig)
	SetupUtility()
	SetupController()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for sig := range sigCh {
			rootLogger.Sugar().Warnf("Received signal %+v ", sig)
			os.Exit(1)
		}
	}()
	select {}
}
