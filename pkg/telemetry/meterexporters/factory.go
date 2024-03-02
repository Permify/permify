package meterexporters

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/metric"
)

// ExporterFactory - Create meter exporter according to given params
func ExporterFactory(name, endpoint string, insecure bool, urlpath string, grpc bool) (metric.Exporter, error) {
	switch name {
	case "otlp":
		if grpc {
			return NewOTLPGrpc(endpoint, insecure)
		}
		return NewOTLPHttp(endpoint, insecure, urlpath)
	default:
		return nil, fmt.Errorf("%s meter exporter is unsupported", name)
	}
}
