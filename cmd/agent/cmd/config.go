// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/elf-io/elf/pkg/logger"
	"github.com/elf-io/elf/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strconv"
)

func init() {

	viper.AutomaticEnv()
	if t := viper.GetString("ENV_LOG_LEVEL"); len(t) > 0 {
		rootLogger = logger.NewStdoutLogger(t, BinName).Named(BinName)
		rootLogger.Info("ENV_LOG_LEVEL = " + t)
	} else {
		rootLogger = logger.NewStdoutLogger("", BinName).Named(BinName)
		rootLogger.Info("ENV_LOG_LEVEL is empty ")
	}

	logger := rootLogger.Named("config")
	// env built in the image
	if t := viper.GetString("ENV_VERSION"); len(t) > 0 {
		logger.Info("app version " + t)
	}
	if t := viper.GetString("ENV_GIT_COMMIT_VERSION"); len(t) > 0 {
		logger.Info("git commit version " + t)
	}
	if t := viper.GetString("ENV_GIT_COMMIT_TIMESTAMP"); len(t) > 0 {
		logger.Info("git commit timestamp " + t)
	}

	for n, v := range types.AgentEnvMapping {
		m := v.DefaultValue
		if t := viper.GetString(v.EnvName); len(t) > 0 {
			m = t
		}
		if len(m) > 0 {
			switch v.DestVar.(type) {
			case *int32:
				if s, err := strconv.ParseInt(m, 10, 64); err == nil {
					r := types.AgentEnvMapping[n].DestVar.(*int32)
					*r = int32(s)
				} else {
					logger.Fatal("failed to parse env value of " + v.EnvName + " to int32, value=" + m)
				}
			case *string:
				r := types.AgentEnvMapping[n].DestVar.(*string)
				*r = m
			case *bool:
				if s, err := strconv.ParseBool(m); err == nil {
					r := types.AgentEnvMapping[n].DestVar.(*bool)
					*r = s
				} else {
					logger.Fatal("failed to parse env value of " + v.EnvName + " to bool, value=" + m)
				}
			default:
				logger.Sugar().Fatal("unsupported type to parse %v, config type=%v ", v.EnvName, reflect.TypeOf(v.DestVar))
			}
		}

		logger.Info(v.EnvName + " = " + m)
	}

	if len(types.AgentConfig.LocalNodeName) == 0 {
		types.AgentConfig.LocalNodeName, _ = os.Hostname()
	}

	// command flags
	globalFlag := rootCmd.PersistentFlags()
	globalFlag.StringVarP(&types.AgentConfig.ConfigMapPath, "config-path", "C", "", "configmap file path")
	if e := viper.BindPFlags(globalFlag); e != nil {
		logger.Sugar().Fatalf("failed to BindPFlags, reason=%v", e)
	}
	printFlag := func() {
		logger.Info("config-path = " + types.AgentConfig.ConfigMapPath)

		// load configmap
		if len(types.AgentConfig.ConfigMapPath) > 0 {
			configmapBytes, err := os.ReadFile(types.AgentConfig.ConfigMapPath)
			if nil != err {
				logger.Sugar().Fatalf("failed to read configmap file %v, error: %v", types.AgentConfig.ConfigMapPath, err)
			}
			if err := yaml.Unmarshal(configmapBytes, &types.AgentConfig.Configmap); nil != err {
				logger.Sugar().Fatalf("failed to parse configmap data, error: %v", err)
			}
		}
	}
	cobra.OnInitialize(printFlag)

}
