// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package debug

import (
	"fmt"
	"github.com/google/gops/agent"
	pyroscope "github.com/grafana/pyroscope-go"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"runtime"
)

type DebugManager interface {
	RunGops(port int)
	RunPyroscope(serverAddress string, localHostName string)
}

type debugManager struct {
	logger *zap.Logger
}

var _ DebugManager = (*debugManager)(nil)

func (s *debugManager) RunGops(listerPort int) {
	address := fmt.Sprintf("127.0.0.1:%d", listerPort)
	op := agent.Options{
		ShutdownCleanup: true,
		Addr:            address,
	}
	if err := agent.Listen(op); err != nil {
		s.logger.Sugar().Fatalf("gops failed to listen on port %s, reason=%v", address, err)
	}
	s.logger.Sugar().Infof("gops is listening on %s ", address)
	// defer agent.Close()
}

func (s *debugManager) RunPyroscope(serverAddress string, localHostName string) {
	// push mode ,  push to pyroscope server
	s.logger.Sugar().Infof("%v pyroscope works in push mode, server %s ", localHostName, serverAddress)

	// These 2 lines are only required if you're using mutex or block profiling
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	_, e := pyroscope.Start(pyroscope.Config{
		ApplicationName: filepath.Base(os.Args[0]),
		ServerAddress:   serverAddress,
		// too much log
		// Logger:          pyroscope.StandardLogger,
		Logger: nil,
		Tags:   map[string]string{"node": localHostName},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if e != nil {
		s.logger.Sugar().Fatalf("failed to setup pyroscope, reason=%v", e)
	}
}

func New(logger *zap.Logger) DebugManager {
	return &debugManager{
		logger: logger,
	}
}
