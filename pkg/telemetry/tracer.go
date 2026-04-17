package telemetry

import (
	"context"
	"os"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation" // Trace context propagation
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	"github.com/Permify/permify/internal"
)

// NewTracer - Creates new tracer
func NewTracer(exporter trace.SpanExporter, serviceName string) func(context.Context) error {
	if strings.TrimSpace(serviceName) == "" {
		serviceName = "permify"
	}
	hostName, err := os.Hostname()
	if err != nil {
		return func(context.Context) error { return nil }
	}

	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exporter)),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("id", internal.Identifier),
			attribute.String("project.id", internal.Identifier),
			attribute.String("version", internal.Version),
			attribute.String("host_name", hostName),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{}) // Set trace context propagator
	// Return shutdown function
	return tp.Shutdown
}
