// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/elf-io/balancing/pkg/debug"
	"github.com/elf-io/balancing/pkg/types"
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

	rootLogger.Sugar().Infof("config: %+v", types.ControllerConfig)

	SetupUtility()

	SetupController()

}
