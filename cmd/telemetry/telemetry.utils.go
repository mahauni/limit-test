package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type Telemetry struct {
	ctx               context.Context
	collectorEndpoint string
	serviceName       string
	environment       string
}

func NewTelemetry(ctx context.Context, collectorEndpoint string, serviceName string, environment string) *Telemetry {
	return &Telemetry{
		ctx,
		collectorEndpoint,
		serviceName,
		environment,
	}
}

func (t *Telemetry) InitTracerProvider() (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(t.ctx,
		otlptracehttp.WithEndpoint(t.collectorEndpoint),
		otlptracehttp.WithInsecure(), // Use WithInsecure for local testing with HTTP
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(t.serviceName),
			attribute.String("environment", t.environment),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func (t *Telemetry) InitMeterProvider() (*sdkmetric.MeterProvider, error) {
	exporter, err := otlpmetrichttp.New(t.ctx,
		otlpmetrichttp.WithEndpoint(t.collectorEndpoint),
		otlpmetrichttp.WithInsecure(), // Use WithInsecure for local testing with HTTP
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(3*time.Second))),
		sdkmetric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(t.serviceName),
			attribute.String("environment", t.environment),
		)),
	)
	otel.SetMeterProvider(mp)
	return mp, nil
}
