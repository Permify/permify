package logexporters

import (
	"context"
	"fmt"

	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogsgrpc"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogshttp"
)

// NewOTLP - Creates new OTLP exporter based on protocol.
func NewOTLP(endpoint string, insecure bool, urlpath string, headers map[string]string, protocol string) (*otlplogs.Exporter, error) {
	switch protocol {
	case "http":
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

	case "grpc":
		options := []otlplogsgrpc.Option{
			otlplogsgrpc.WithEndpoint(endpoint),
		}

		if insecure {
			options = append(options, otlplogsgrpc.WithInsecure())
		}

		exporter, err := otlplogs.NewExporter(context.Background(), otlplogs.WithClient(
			otlplogsgrpc.NewClient(options...),
		))
		if err != nil {
			return nil, err
		}

		return exporter, nil

	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
