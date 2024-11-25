package otel

import (
	"context"
)

func Start(ctx context.Context, serviceName string) (ShutdownFunc, error) {
	tracing, err := startTracerProvider(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	metrics, err := startMetricsProvider(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	return func() error {
		if err := tracing.Shutdown(ctx); err != nil {
			return err
		}

		if err := metrics.Shutdown(ctx); err != nil {
			return err
		}

		return nil
	}, nil
}
