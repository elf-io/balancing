// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"github.com/elf-io/elf/pkg/debug"
	"github.com/elf-io/elf/pkg/mybookManager"
	"github.com/elf-io/elf/pkg/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"path/filepath"
	"time"
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

	SetupHttpServer()

	// ------

	RunMetricsServer(types.ControllerConfig.PodName)
	MetricGaugeEndpoint.Add(context.Background(), 100)
	MetricGaugeEndpoint.Add(context.Background(), -10)
	MetricGaugeEndpoint.Add(context.Background(), 5)

	attrs := attribute.NewSet(attribute.String("pod1", "value1"), attribute.Int("version", 1))
	MetricCounterRequest.Add(context.Background(), 10, metric.WithAttributeSet(attrs))

	attrs = attribute.NewSet(attribute.String("pod2", "value1"), attribute.Int("version", 1))
	MetricCounterRequest.Add(context.Background(), 5, metric.WithAttributeSet(attrs))

	MetricHistogramDuration.Record(context.Background(), 10)
	MetricHistogramDuration.Record(context.Background(), 20)

	// ----------
	s := mybookManager.New(rootLogger.Named("mybook"))
	s.RunController("testlease", types.ControllerConfig.PodNamespace, types.ControllerConfig.PodName)
	s.RunWebhookServer(int(types.ControllerConfig.WebhookPort), filepath.Dir(types.ControllerConfig.TlsServerCertPath))

	// ------------
	rootLogger.Info("hello world")
	time.Sleep(time.Hour)
}
