package meterexporters

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

// NewOTLP - Creates new OTLP exporter
func NewOTLP(endpoint string) (metric.Exporter, error) {
	exporter, err := otlpmetrichttp.New(context.Background(),
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
		otlpmetrichttp.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}
