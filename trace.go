package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
)

var tracer = otel.Tracer("oyaki")

func initProvider(ctx context.Context) func() {
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("oyaki"),
		),
	)
	handleErr(err, "failed to create resource")

	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithDialOption(grpc.WithBlock()))
	traceExp, err := otlptrace.New(ctx, traceClient)
	handleErr(err, "Failed to create the collector trace exporter")

	samplingRate := os.Getenv("OTEL_TRACES_SAMPLING_RATE")
	if samplingRate == "" {
		samplingRate = "1"
	}
	samplingRateFloat, err := strconv.ParseFloat(samplingRate, 64)
	if err != nil {
		handleErr(err, "Failed to parse OTEL_TRACES_SAMPLING_RATE as float64, using default 1.0")
	}
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplingRateFloat))

	bsp := sdktrace.NewBatchSpanProcessor(traceExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracerProvider)

	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := traceExp.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}
}

func handleErr(err error, message string) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}
