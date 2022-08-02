package exporters

import (
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/trace"
)

// NewZipkin -
func NewZipkin(url string) (trace.SpanExporter, error) {
	exporter, err := zipkin.New(url)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}
