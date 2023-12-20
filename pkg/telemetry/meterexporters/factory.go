package meterexporters

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/metric"
)

// ExporterFactory - Create meter exporter according to given params
func ExporterFactory(name, endpoint string) (metric.Exporter, error) {
	switch name {
	case "otlp":
		return NewOTLP(endpoint)
	default:
		return nil, fmt.Errorf("%s meter exporter is unsupported", name)
	}
}
