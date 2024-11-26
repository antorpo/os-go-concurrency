package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

const (
	_collectPeriod   = 30 * time.Second
	_minimumInterval = time.Minute
)

var _histogramBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10, 25, 50, 100}

func startMetricsProvider(ctx context.Context, serviceName string) (*metric.MeterProvider, error) {
	exp, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(getEndpoint()),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	resources := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
	)

	mp := metric.NewMeterProvider(
		metric.WithResource(resources),
		metric.WithReader(
			metric.NewPeriodicReader(
				exp,
				metric.WithInterval(_collectPeriod))),
		metric.WithView(metric.NewView(
			metric.Instrument{
				Name: "*",
				Kind: metric.InstrumentKindHistogram,
			},
			metric.Stream{
				Aggregation: metric.AggregationExplicitBucketHistogram{
					Boundaries: _histogramBuckets,
				},
			},
		)),
	)
	otel.SetMeterProvider(mp)

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(_minimumInterval))
	if err != nil {
		return nil, err
	}

	return mp, nil
}
