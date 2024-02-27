package meterexporters

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/metric"
)

// ExporterFactory - Create meter exporter according to given params
func ExporterFactory(name, endpoint string, insecure bool, urlpath string) (metric.Exporter, error) {
	switch name {
	case "otlp":
		return NewOTLPHttp(endpoint, insecure, urlpath)
	case "otlp-grpc":
		return NewOTLPGrpc(endpoint, insecure)
	default:
		return nil, fmt.Errorf("%s meter exporter is unsupported", name)
	}
}
