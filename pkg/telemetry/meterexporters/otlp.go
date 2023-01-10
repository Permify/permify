package meterexporters

import (
	"context"
	`google.golang.org/grpc/encoding/gzip`
	`time`

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
)

// NewOTLP - Creates new OTLP exporter
func NewOTLP(endpoint string) (metric.Exporter, error) {
	exporter, err := otlpmetricgrpc.New(context.Background(),
		otlpmetricgrpc.WithCompressor(gzip.Name),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithReconnectionPeriod(2*time.Second),
		otlpmetricgrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}
