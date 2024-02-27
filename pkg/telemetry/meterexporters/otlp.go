package meterexporters

import (
	"context"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

const GRPC_COMPRESSION_STRING = "gzip"

// NewOTLPHttp - Creates new OTLP exporter
func NewOTLPHttp(endpoint string, insecure bool, urlpath string) (metric.Exporter, error) {
	options := []otlpmetrichttp.Option{
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
		otlpmetrichttp.WithEndpoint(endpoint),
	}

	if urlpath != "" {
		options = append(options, otlpmetrichttp.WithURLPath(urlpath))
	}

	if insecure {
		options = append(options, otlpmetrichttp.WithInsecure())
	}

	exporter, err := otlpmetrichttp.New(context.Background(), options...)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}

func NewOTLPGrpc(endpoint string, insecure bool) (metric.Exporter, error) {
	options := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithCompressor(GRPC_COMPRESSION_STRING),
		otlpmetricgrpc.WithEndpoint(endpoint),
	}

	if insecure {
		options = append(options, otlpmetricgrpc.WithInsecure())
	}

	exporter, err := otlpmetricgrpc.New(context.Background(), options...)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
