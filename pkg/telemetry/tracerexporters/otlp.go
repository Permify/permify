package tracerexporters

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

// NewOTLP - Creates new OTLP exporter
func NewOTLP(url string) (trace.SpanExporter, error) {
	var exporter trace.SpanExporter
	var err error
	exporter, err = otlptracehttp.New(context.Background(), otlptracehttp.WithInsecure(), otlptracehttp.WithEndpoint(url))
	if err != nil {
		return nil, err
	}
	return exporter, nil
}
