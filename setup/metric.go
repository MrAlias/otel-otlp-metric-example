package setup

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

func NewMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {
	exp, err := NewHTTPExporter(ctx)
	if err != nil {
		return nil, err
	}

	res, err := NewResource(ctx)
	if err != nil {
		return nil, err
	}

	// Set the reader collection periord to 10 seconds (default 60).
	reader := metric.NewPeriodicReader(exp, metric.WithInterval(10*time.Second))
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(reader),
		metric.WithView(metric.NewView(
			// Have any instrument, from any instrumentation library, with the name
			// "request.duration" use these buckets.
			metric.Instrument{Name: "request.duration"},
			metric.Stream{
				Aggregation: aggregation.ExplicitBucketHistogram{
					Boundaries: []float64{0.000001, 0.00001, 0.0001, 0.001, 0.01, 0.1, 1, 10},
				},
			},
		)),
	)

	return meterProvider, nil
}

func NewGRPCExporter(ctx context.Context) (metric.Exporter, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// Exponential back-off strategy.
	backoffConf := backoff.DefaultConfig
	// You can also change the base delay, multiplier, and jitter here.
	backoffConf.MaxDelay = 240 * time.Second

	conn, err := grpc.DialContext(
		ctx,
		"127.0.0.1:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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

func NewHTTPExporter(ctx context.Context) (metric.Exporter, error) {
	return otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithInsecure(),
		// WithTimeout sets the max amount of time the Exporter will attempt an
		// export.
		otlpmetrichttp.WithTimeout(7*time.Second),
		otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{
			// Enabled indicates whether to not retry sending batches in case
			// of export failure.
			Enabled: true,
			// InitialInterval the time to wait after the first failure before
			// retrying.
			InitialInterval: 1 * time.Second,
			// MaxInterval is the upper bound on backoff interval. Once this
			// value is reached the delay between consecutive retries will
			// always be `MaxInterval`.
			MaxInterval: 10 * time.Second,
			// MaxElapsedTime is the maximum amount of time (including retries)
			// spent trying to send a request/batch. Once this value is
			// reached, the data is discarded.
			MaxElapsedTime: 240 * time.Second,
		}),
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
