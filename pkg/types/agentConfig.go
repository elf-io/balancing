// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package types

type AgentConfigStruct struct {
	// --------- from env
	EnableMetric           bool
	MetricPort             int32
	HttpPort               int32
	GopsPort               int32
	PyroscopeServerAddress string
	PodName                string
	PodNamespace           string
	LocalNodeName          string
	EbpfLogLevel           string

	// ------------- from flags
	ConfigMapPath     string
	TlsCaCertPath     string
	TlsServerCertPath string
	TlsServerKeyPath  string

	// ------------ from configmap
	Configmap ConfigmapConfig
}

var AgentConfig AgentConfigStruct

var AgentEnvMapping = []EnvMapping{
	{"ENV_ENABLED_METRIC", "false", &AgentConfig.EnableMetric},
	{"ENV_METRIC_HTTP_PORT", "", &AgentConfig.MetricPort},
	{"ENV_HTTP_PORT", "5810", &AgentConfig.HttpPort},
	{"ENV_GOPS_LISTEN_PORT", "", &AgentConfig.GopsPort},
	{"ENV_PYROSCOPE_PUSH_SERVER_ADDRESS", "", &AgentConfig.PyroscopeServerAddress},
	{"ENV_POD_NAME", "", &AgentConfig.PodName},
	{"ENV_POD_NAMESPACE", "", &AgentConfig.PodNamespace},
	{"ENV_LOCAL_NODE_NAME", "", &AgentConfig.LocalNodeName},
	{"ENV_EBPF_LOG_LEVEL", "verbose", &AgentConfig.EbpfLogLevel},
}
