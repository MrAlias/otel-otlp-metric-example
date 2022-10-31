package setup

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/view"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
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

	return otlpmetricgrpc.New(ctx,
		// Default collector endpoint
		otlpmetricgrpc.WithEndpoint("127.0.0.1:4317"),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithDialOption(grpc.WithBlock()),
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

func otelMetricsStart(exporter metric.Exporter, res *resource.Resource) *metric.MeterProvider {
	meterProv := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	)

	global.SetMeterProvider(meterProv)

	return meterProv
}
