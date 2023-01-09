package meterexporters

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
)

// NewOTLP - Creates new OTLP exporter
func NewOTLP(endpoint string) (metric.Exporter, error) {
	exporter, err := otlpmetricgrpc.New(context.Background(), otlpmetricgrpc.WithInsecure(), otlpmetricgrpc.WithEndpoint(endpoint))
	if err != nil {
		return nil, err
	}
	return exporter, nil
}
