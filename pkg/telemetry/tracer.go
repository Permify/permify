package telemetry

import (
	"context"
	"runtime"

	"github.com/rs/xid"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.10.0"

	`github.com/Permify/permify/internal`
)

// NewTracer - Creates new tracer
func NewTracer(exporter trace.SpanExporter) func(context.Context) error {
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exporter)),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("permify"),
			attribute.String("id", xid.New().String()),
			attribute.String("version", internal.Version),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown
}
