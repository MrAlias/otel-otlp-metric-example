# otel-otlp-metric-example

Example of how to setup an OTLP exporter for metrics

## Run

Start the collector using docker-compose:

```terminal
docker-compose up -d
```

Run the client:

```terminal
go run .
```

Make an HTTP request to the Go App:

```terminal
curl -v localhost:8080/
```

Wait 10 seconds for the periodic reader collection cycle to expire and watch the logs for data:

```terminal
docker-compose logs -f
```

If everything goes right, you should see the collector logging the histogram data it has received.

```
...
otel-otlp-metric-example-otel-collector-1  | 2022-10-31T19:50:40.376Z	INFO	loggingexporter/logging_exporter.go:56	MetricsExporter	{"#metrics": 1}
otel-otlp-metric-example-otel-collector-1  | 2022-10-31T19:50:40.376Z	DEBUG	loggingexporter/logging_exporter.go:66	ResourceMetrics #0
otel-otlp-metric-example-otel-collector-1  | Resource labels:
otel-otlp-metric-example-otel-collector-1  |      -> env: STRING(dev)
otel-otlp-metric-example-otel-collector-1  |      -> service.instance.id: STRING(xi)
otel-otlp-metric-example-otel-collector-1  |      -> service.name: STRING(serviceName)
otel-otlp-metric-example-otel-collector-1  |      -> telemetry.sdk.language: STRING(go)
otel-otlp-metric-example-otel-collector-1  |      -> telemetry.sdk.name: STRING(opentelemetry)
otel-otlp-metric-example-otel-collector-1  |      -> telemetry.sdk.version: STRING(1.11.1)
otel-otlp-metric-example-otel-collector-1  | InstrumentationLibraryMetrics #0
otel-otlp-metric-example-otel-collector-1  | InstrumentationLibrary go.opentelemetry.io/otel/example/otlp v0.1.1
otel-otlp-metric-example-otel-collector-1  | Metric #0
otel-otlp-metric-example-otel-collector-1  | Descriptor:
otel-otlp-metric-example-otel-collector-1  |      -> Name: request.duration
otel-otlp-metric-example-otel-collector-1  |      -> Description: Time taken to perfrom a user request
otel-otlp-metric-example-otel-collector-1  |      -> Unit: ms
otel-otlp-metric-example-otel-collector-1  |      -> DataType: Histogram
otel-otlp-metric-example-otel-collector-1  |      -> AggregationTemporality: AGGREGATION_TEMPORALITY_CUMULATIVE
otel-otlp-metric-example-otel-collector-1  | HistogramDataPoints #0
otel-otlp-metric-example-otel-collector-1  | StartTimestamp: 2022-10-31 19:50:10.268993381 +0000 UTC
otel-otlp-metric-example-otel-collector-1  | Timestamp: 2022-10-31 19:50:40.273494074 +0000 UTC
otel-otlp-metric-example-otel-collector-1  | Count: 1
otel-otlp-metric-example-otel-collector-1  | Sum: 0.000000
otel-otlp-metric-example-otel-collector-1  | ExplicitBounds #0: 0.010000
otel-otlp-metric-example-otel-collector-1  | ExplicitBounds #1: 0.100000
otel-otlp-metric-example-otel-collector-1  | ExplicitBounds #2: 1.000000
otel-otlp-metric-example-otel-collector-1  | ExplicitBounds #3: 10.000000
otel-otlp-metric-example-otel-collector-1  | ExplicitBounds #4: 100.000000
otel-otlp-metric-example-otel-collector-1  | ExplicitBounds #5: 1000.000000
otel-otlp-metric-example-otel-collector-1  | Buckets #0, Count: 1
otel-otlp-metric-example-otel-collector-1  | Buckets #1, Count: 0
otel-otlp-metric-example-otel-collector-1  | Buckets #2, Count: 0
otel-otlp-metric-example-otel-collector-1  | Buckets #3, Count: 0
otel-otlp-metric-example-otel-collector-1  | Buckets #4, Count: 0
otel-otlp-metric-example-otel-collector-1  | Buckets #5, Count: 0
otel-otlp-metric-example-otel-collector-1  | Buckets #6, Count: 0
...
```
