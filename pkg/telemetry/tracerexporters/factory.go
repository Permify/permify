package tracerexporters

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/trace"
)

// ExporterFactory - Create tracer exporter according to given params
func ExporterFactory(name, url string) (trace.SpanExporter, error) {
	switch name {
	case "zipkin":
		return NewZipkin(url)
	case "jaeger":
		return NewJaegar(url)
	case "otlp":
		return NewOTLP(url)
	case "signoz":
		return NewSigNoz(url, false)
	default:
		return nil, fmt.Errorf("%s tracer exporter is unsupported", name)
	}
}
