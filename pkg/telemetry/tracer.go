package telemetry

import (
	"context"
	"os"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	"github.com/Permify/permify/internal"
)

// NewTracer - Creates new tracer
func NewTracer(exporter trace.SpanExporter) func(context.Context) error {
	hostName, err := os.Hostname()
	if err != nil {
		return func(context.Context) error { return nil }
	}

	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exporter)),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("permify"),
			attribute.String("id", internal.Identifier),
			attribute.String("project.id", internal.Identifier),
			attribute.String("version", internal.Version),
			attribute.String("host_name", hostName),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown
}
