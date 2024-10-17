// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package types

type ControllerConfigStruct struct {
	// ---------- from env
	EnableMetric           bool
	MetricPort             int32
	HttpPort               int32
	GopsPort               int32
	WebhookPort            int32
	PyroscopeServerAddress string
	PodName                string
	PodNamespace           string

	// ----------  from flags
	ConfigMapPath     string
	TlsCaCertPath     string
	TlsServerCertPath string
	TlsServerKeyPath  string

	// ----------  from configmap
	Configmap ConfigmapConfig
}

var ControllerConfig ControllerConfigStruct

var ControllerEnvMapping = []EnvMapping{
	{"ENV_ENABLED_METRIC", "false", &ControllerConfig.EnableMetric},
	{"ENV_METRIC_HTTP_PORT", "", &ControllerConfig.MetricPort},
	{"ENV_HTTP_PORT", "5720", &ControllerConfig.HttpPort},
	{"ENV_GOPS_LISTEN_PORT", "", &ControllerConfig.GopsPort},
	{"ENV_WEBHOOK_PORT", "", &ControllerConfig.WebhookPort},
	{"ENV_PYROSCOPE_PUSH_SERVER_ADDRESS", "", &ControllerConfig.PyroscopeServerAddress},
	{"ENV_POD_NAME", "", &ControllerConfig.PodName},
	{"ENV_POD_NAMESPACE", "", &ControllerConfig.PodNamespace},
}
