package meterexporters

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

// NewOTLP - Creates new OTLP exporter using HTTP protocol.
func NewOTLP(endpoint string, insecure bool, urlpath string, headers map[string]string) (metric.Exporter, error) {
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
}
