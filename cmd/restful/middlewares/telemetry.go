package middlewares

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type OtelTelemetry struct {
	meter string
	*sdktrace.TracerProvider
	*sdkmetric.MeterProvider
}

func NewOtelTelemtry(trace *sdktrace.TracerProvider, metric *sdkmetric.MeterProvider, meter string) OtelTelemetry {
	return OtelTelemetry{
		meter,
		trace,
		metric,
	}
}

func (t *OtelTelemetry) Telemetry(next http.Handler) http.Handler {
	meter := t.MeterProvider.Meter(t.meter)
	HTTPServerDuration := "dur"
	HTTPServerActiveRequests := "act-req"
	HTTPServerRequestSize := "req-size"
	HTTPServerResponseSize := "res-size"
	metrics := []struct {
		name        string
		description string
		register    func(metric.Meter) (any, error)
	}{
		{HTTPServerDuration, "Measures the duration of inbound HTTP requests", func(m metric.Meter) (any, error) { return m.Float64Histogram(HTTPServerDuration) }},
		{HTTPServerActiveRequests, "Number of concurrent HTTP requests in flight", func(m metric.Meter) (any, error) { return m.Int64UpDownCounter(HTTPServerActiveRequests) }},
		{HTTPServerRequestSize, "Measures the size of HTTP request messages", func(m metric.Meter) (any, error) { return m.Float64Histogram(HTTPServerRequestSize) }},
		{HTTPServerResponseSize, "Measures the size of HTTP response messages", func(m metric.Meter) (any, error) { return m.Float64Histogram(HTTPServerResponseSize) }},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, metricDef := range metrics {
			if _, err := metricDef.register(meter); err != nil {
				fmt.Printf("[Telemetry] Failed to create metric %s: %v\n", metricDef.name, err)
			}
		}
	})
}
