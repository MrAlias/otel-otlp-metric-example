package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/MrAlias/otel-otlp-metric-example/setup"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
)

const (
	// instName needs to be the name of the package providing instrumentation.
	instName = "go.opentelemetry.io/otel/example/otlp"
	instVer  = "v0.1.1"
)

type App struct {
	MeterProvider metric.MeterProvider

	reqDuration syncint64.Histogram
}

func NewApp(mp metric.MeterProvider) (*App, error) {
	app := &App{MeterProvider: mp}

	meter := mp.Meter(instName, metric.WithInstrumentationVersion(instVer))
	var err error
	app.reqDuration, err = meter.Int64Histogram(
		"request.duration",
		instrument.WithDescription("Time taken to perfrom a user request"),
		instrument.WithUnit(unit.Milliseconds),
	)
	return app, err
}

func (a *App) Run(addr string) {
	log.Printf("serving metrics at %s/", addr)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defer func(start time.Time) {
			d := time.Since(start).Milliseconds()
			a.reqDuration.Record(req.Context(), d)
		}(time.Now())
		w.WriteHeader(http.StatusOK)
	})
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Println(err)
	}
}

func main() {
	ctx := context.Background()
	meterProvider, err := setup.NewMeterProvider(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	app, err := NewApp(meterProvider)
	if err != nil {
		log.Fatalln(err)
	}

	go app.Run(":8080")

	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
	<-ctx.Done()
}
