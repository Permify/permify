package meterexporters

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/metric"
)

// ExporterFactory - Create meter exporter according to given params
func ExporterFactory(name, endpoint string, insecure bool, urlpath string, headers map[string]string, protocol string) (metric.Exporter, error) {
	switch name {
	case "otlp", "otlp-http", "otlp-grpc":
		return NewOTLP(endpoint, insecure, urlpath, headers, protocol)
	case "gcp":
		return NewGCP(headers)
	default:
		return nil, fmt.Errorf("%s meter exporter is unsupported", name)
	}
}
