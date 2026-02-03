// Package telemetry provides OpenTelemetry instrumentation for Honeycomb.
package telemetry

import (
	"context"
	"os"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	serviceName    = "dungeonband"
	serviceVersion = "0.1.0"
)

// Setup initializes OpenTelemetry with OTLP HTTP exporter.
// It reads configuration from standard OTEL_* environment variables:
//   - OTEL_EXPORTER_OTLP_ENDPOINT: Honeycomb endpoint (https://api.honeycomb.io)
//   - OTEL_EXPORTER_OTLP_HEADERS: Headers including x-honeycomb-team=<api-key>
//
// Returns a shutdown function that should be called on application exit.
func Setup(ctx context.Context) (shutdown func(context.Context) error, err error) {
	// Create OTLP HTTP exporter - automatically uses OTEL_* env vars
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	// Build resource with service information
	// We create our own resource without merging with Default() to avoid schema URL conflicts
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("service.version", serviceVersion),
			attribute.String("telemetry.sdk.language", "go"),
			attribute.String("telemetry.sdk.name", "opentelemetry"),
			attribute.String("host.name", getHostname()),
			attribute.String("os.type", runtime.GOOS),
			attribute.String("process.runtime.name", "go"),
			attribute.String("process.runtime.version", runtime.Version()),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create trace provider with batch span processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Register as global provider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}

// Tracer returns a named tracer for the given component.
// Use this to create spans within different parts of the application.
func Tracer(name string) trace.Tracer {
	return otel.GetTracerProvider().Tracer("dungeonband/" + name)
}

// NoopTracer returns a no-op tracer for use when telemetry is disabled.
func NoopTracer() trace.Tracer {
	return noop.NewTracerProvider().Tracer("dungeonband/noop")
}

// getHostname returns the system hostname, or "unknown" if it cannot be determined.
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
