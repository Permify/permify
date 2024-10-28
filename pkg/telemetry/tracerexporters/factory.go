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
	case "otlp", "otlp-http", "otlp-grpc":
		return NewOTLP(url, insecure, urlpath, headers, protocol)
	case "signoz":
		return NewSigNoz(url, insecure, headers)
	case "gcp":
		return NewGCP(headers)
	default:
		return nil, fmt.Errorf("%s tracer exporter is unsupported", name)
	}
}
