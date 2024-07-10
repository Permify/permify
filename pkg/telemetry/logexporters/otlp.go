package logexporters

import (
	"context"

	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogshttp"
)

// NewOTLP - Creates new OTLP exporter using HTTP protocol.
func NewOTLP(endpoint string, insecure bool, urlpath string, headers map[string]string) (*otlplogs.Exporter, error) {
	options := []otlplogshttp.Option{
		otlplogshttp.WithCompression(otlplogshttp.GzipCompression),
		otlplogshttp.WithEndpoint(endpoint),
	}

	if urlpath != "" {
		options = append(options, otlplogshttp.WithURLPath(urlpath))
	}

	if insecure {
		options = append(options, otlplogshttp.WithInsecure())
	}

	exporter, err := otlplogs.NewExporter(context.Background(), otlplogs.WithClient(
		otlplogshttp.NewClient(options...),
	))
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
