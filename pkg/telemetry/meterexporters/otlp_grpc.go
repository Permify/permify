package meterexporters

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc/credentials"
)

// NewOTLPGrpc - Creates new OTLP exporter using GRPC protocol.
func NewOTLPGrpc(endpoint string, insecure bool, headers map[string]string) (metric.Exporter, error) {
	options := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithHeaders(headers),
	}

	if len(headers) > 0 {
		options = append(options, otlpmetricgrpc.WithHeaders(headers))
	}

	if insecure {
		options = append(options, otlpmetricgrpc.WithInsecure())
	} else {
		options = append(options, otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	exporter, err := otlpmetricgrpc.New(context.Background(), options...)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
