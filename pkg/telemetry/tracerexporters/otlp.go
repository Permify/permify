package tracerexporters

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// NewOTLP - Creates new OTLP exporter based on protocol.
func NewOTLP(endpoint string, insecure bool, urlpath string, headers map[string]string, protocol string) (trace.SpanExporter, error) {
	switch protocol {
	case "http":
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(endpoint),
		}

		if len(headers) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(headers))
		}

		if urlpath != "" {
			opts = append(opts, otlptracehttp.WithURLPath(urlpath))
		}

		if insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}

		exporter, err := otlptracehttp.New(context.Background(), opts...)
		if err != nil {
			return nil, err
		}

		return exporter, nil

	case "grpc":
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(endpoint),
		}

		if len(headers) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(headers))
		}

		if insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		} else {
			opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
		}

		exporter, err := otlptracegrpc.New(context.Background(), opts...)
		if err != nil {
			return nil, err
		}

		return exporter, nil

	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
