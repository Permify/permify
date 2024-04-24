package tracerexporters

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

// NewOTLP - Creates new OTLP exporter using HTTP protocol.
func NewOTLP(endpoint string, insecure bool, urlpath string) (trace.SpanExporter, error) {
	var exporter trace.SpanExporter
	var err error

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
	}

	if urlpath != "" {
		opts = append(opts, otlptracehttp.WithURLPath(urlpath))
	}

	if insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err = otlptracehttp.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
