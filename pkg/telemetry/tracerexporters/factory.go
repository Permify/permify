package tracerexporters

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/trace"
)

// ExporterFactory - Create tracer exporter according to given params
func ExporterFactory(name, url string, insecure bool, urlpath string, headers map[string]string, protocol string) (trace.SpanExporter, error) {
	switch name {
	case "zipkin":
		return NewZipkin(url)
	case "jaeger":
		return NewJaegar(url)
	case "otlp", "otlp-http":
		return NewOTLP(url, insecure, urlpath, headers, "http")
	case "otlp-grpc":
		return NewOTLP(url, insecure, urlpath, headers, "grpc")
	case "signoz":
		return NewSigNoz(url, insecure, headers)
	default:
		return nil, fmt.Errorf("%s tracer exporter is unsupported", name)
	}
}
