package setup

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/view"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

func NewMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {
	exp, err := NewOTLPExporter(ctx)
	if err != nil {
		return nil, err
	}

	res, err := NewResource(ctx)
	if err != nil {
		return nil, err
	}

	views, err := NewViews()
	if err != nil {
		return nil, err
	}

	// Set the reader collection periord to 10 seconds (default 60).
	reader := metric.NewPeriodicReader(exp, metric.WithInterval(10*time.Second))
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(reader, views...),
	)

	return meterProvider, nil
}

func NewViews() ([]view.View, error) {
	views := make([]view.View, 0, 2)

	v, err := view.New(
		// Have any instrument, from any instrumentation library, with the name
		// "request.duration" use these buckets.
		view.MatchInstrumentName("request.duration"),
		view.WithSetAggregation(aggregation.ExplicitBucketHistogram{
			Boundaries: []float64{0.01, 0.1, 1, 10, 100, 1000},
		}),
	)
	if err != nil {
		return nil, err
	}
	views = append(views, v)

	return views, err
}

func NewOTLPExporter(ctx context.Context) (metric.Exporter, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// Exponential back-off strategy.
	backoffConf := backoff.DefaultConfig
	// You can also change the base delay, multiplier, and jitter here.
	backoffConf.MaxDelay = 240 * time.Second

	conn, err := grpc.DialContext(
		ctx,
		"127.0.0.1:4317",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoffConf,
			// Connection timeout.
			MinConnectTimeout: 5 * time.Second,
		}),
	)
	if err != nil {
		return nil, err
	}

	return otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithGRPCConn(conn),
		// WithTimeout sets the max amount of time the Exporter will attempt an
		// export.
		otlpmetricgrpc.WithTimeout(7*time.Second),
	)
}

func NewResource(context.Context) (*resource.Resource, error) {
	base := resource.Default()

	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	return resource.Merge(base, resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("serviceName"),
		semconv.ServiceInstanceIDKey.String(host),
		attribute.String("env", "dev"),
	))
}
