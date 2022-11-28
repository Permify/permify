package exporters

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/sdk/trace"
)

// ExporterFactory - Create tracer according to given params
func ExporterFactory(name string, url string) (trace.SpanExporter, error) {
	switch name {
	case "zipkin":
		return NewZipkin(url)
	case "jaeger":
		return NewJaegar(url)
	case "signoz":
		return NewSigNoz(url, false)
	default:
		return nil, errors.New(fmt.Sprintf("%s exporter is unsupported", name))
	}
}
