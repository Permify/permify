package tracerexporters

import (
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
)

// NewJaegar - Creates new Jaeger exporter
func NewJaegar(url string) (trace.SpanExporter, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	return exporter, nil
}
