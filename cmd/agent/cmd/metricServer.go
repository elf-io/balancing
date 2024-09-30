// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	pkgmetric "github.com/elf-io/elf/pkg/metrics"
	"github.com/elf-io/elf/pkg/types"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	api "go.opentelemetry.io/otel/metric"

	"go.opentelemetry.io/otel/sdk/instrumentation"
)

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

	// globalMeter=pkgmetric.NewMetricsServer(meterName, globalConfig.MetricPort, metricMapping, customBucketsView, logger)
	pkgmetric.RunMetricsServer(types.AgentConfig.EnableMetric, meterName, types.AgentConfig.MetricPort, metricMapping, customBucketsView, logger)

}
