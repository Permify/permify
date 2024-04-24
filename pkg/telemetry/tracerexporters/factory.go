package tracerexporters

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/trace"
)

// ExporterFactory - Create tracer exporter according to given params
func ExporterFactory(name, url string, insecure bool, urlpath string) (trace.SpanExporter, error) {
	switch name {
	case "zipkin":
		return NewZipkin(url)
	case "jaeger":
		return NewJaegar(url)
	case "otlp", "otlp-http":
		return NewOTLP(url, insecure, urlpath)
	case "otlp-grpc":
		return NewOTLPGrpc(url, insecure)
	case "signoz":
		return NewSigNoz(url, insecure)
	default:
		return nil, fmt.Errorf("%s tracer exporter is unsupported", name)
	}
}
