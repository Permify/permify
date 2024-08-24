package meterexporters

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc/credentials"
)

// NewOTLP - Creates new OTLP exporter based on protocol.
func NewOTLP(endpoint string, insecure bool, urlpath string, headers map[string]string, protocol string) (metric.Exporter, error) {
	switch protocol {
	case "http":
		options := []otlpmetrichttp.Option{
			otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
			otlpmetrichttp.WithEndpoint(endpoint),
		}

		if len(headers) > 0 {
			options = append(options, otlpmetrichttp.WithHeaders(headers))
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

	case "grpc":
		options := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(endpoint),
			otlpmetricgrpc.WithHeaders(headers),
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

	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
