// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	pkgmetric "github.com/elf-io/elf/pkg/metrics"
	"github.com/elf-io/elf/pkg/types"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// https://pkg.go.dev/go.opentelemetry.io/otel/metric#hdr-Instruments
/*
All synchronous instruments (Int64Counter, Int64UpDownCounter, Int64Histogram, Float64Counter, Float64UpDownCounter, and Float64Histogram) are used to measure the operation and performance of source code during the source code execution. These instruments only make measurements when the source code they instrument is run.

All asynchronous instruments (Int64ObservableCounter, Int64ObservableUpDownCounter, Int64ObservableGauge, Float64ObservableCounter, Float64ObservableUpDownCounter, and Float64ObservableGauge) are used to measure metrics outside of the execution of source code. They are said to make "observations" via a callback function called once every measurement collection cycle.
*/
var (
	MetricCounterRequest    api.Int64Counter
	MetricGaugeEndpoint     api.Int64UpDownCounter
	MetricHistogramDuration api.Float64Histogram
)

var metricMapping = []pkgmetric.MetricMappingType{
	{P: &MetricCounterRequest, Name: "request_counts", Description: "the request counter"},
	{P: &MetricGaugeEndpoint, Name: "endpoint_number", Description: "the endpoint number"},
	{P: &MetricHistogramDuration, Name: "request_duration_seconds", Description: "the request duration histogram"},
}

// var globalMeter metric.Meter
func RunMetricsServer(meterName string) {
	logger := rootLogger.Named("metric")

	// View to customize histogram buckets
	customBucketsView := sdkmetric.NewView(sdkmetric.Instrument{
		Name:  "*duration*",
		Scope: instrumentation.Scope{Name: meterName},
	}, sdkmetric.Stream{Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
		Boundaries: []float64{1, 10, 20, 50},
	}})

	if customBucketsView == nil {
		logger.Sugar().Fatalf("failed to generate view")
	}

	// globalMeter = pkgmetric.NewMetricsServer(meterName, globalConfig.MetricPort, metricMapping, customBucketsView, logger)
	pkgmetric.RunMetricsServer(types.ControllerConfig.EnableMetric, meterName, types.ControllerConfig.MetricPort, metricMapping, customBucketsView, logger)
}
