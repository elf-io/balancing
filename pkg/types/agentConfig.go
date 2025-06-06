// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package types

type AgentConfigStruct struct {
	// --------- from env
	EnableMetric            bool
	EnableIpv4              bool
	EnableIpv6              bool
	RedirectQosLimit        int32
	MetricPort              int32
	HttpPort                int32
	GopsPort                int32
	PyroscopeServerAddress  string
	PodName                 string
	PodNamespace            string
	LocalNodeName           string
	EbpfLogLevel            string
	LocalNodeEntryInterface string
	KubeconfigPath          string

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
	{"ENV_NODE_ENTRY_INTERFACE_NAME", "", &AgentConfig.LocalNodeEntryInterface},
	{"ENV_EBPF_LOG_LEVEL", "verbose", &AgentConfig.EbpfLogLevel},
	{"ENV_ENABLE_IPV4", "", &AgentConfig.EnableIpv4},
	{"ENV_ENABLE_IPV6", "", &AgentConfig.EnableIpv6},
	{"ENV_REDIRECT_QOS_LIMIT", "10", &AgentConfig.RedirectQosLimit},
	{"KUBECONFIG", "", &AgentConfig.KubeconfigPath},
}
