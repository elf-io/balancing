// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/elf-io/balancing/pkg/debug"
	"github.com/elf-io/balancing/pkg/types"
	"time"
)

func SetupUtility() {
	// run gops
	d := debug.New(rootLogger)
	if types.AgentConfig.GopsPort != 0 {
		d.RunGops(int(types.AgentConfig.GopsPort))
	}

	if types.AgentConfig.PyroscopeServerAddress != "" {
		d.RunPyroscope(types.AgentConfig.PyroscopeServerAddress, types.AgentConfig.PodName)
	}
}

func DaemonMain() {
	rootLogger.Sugar().Infof("config: %+v", types.AgentConfig)

	SetupUtility()

	// SetupHttpServer()

	// RunMetricsServer(types.AgentConfig.PodName)

	// SetupController()
	RunReconciles()

	rootLogger.Info("finish all setup ")
	time.Sleep(time.Hour)

}
