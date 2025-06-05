// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package types

type ConfigmapConfig struct {
	EnableIPv4      bool   `yaml:"enableIPv4"`
	EnableIPv6      bool   `yaml:"enableIPv6"`
	RedirectQosLimit int32  `yaml:"redirectQosLimit"`
	ApiServerHost   string `yaml:"apiServerHost"`
	ApiServerPort   string `yaml:"apiServerPort"`
}

type EnvMapping struct {
	EnvName      string
	DefaultValue string
	DestVar      interface{}
}
