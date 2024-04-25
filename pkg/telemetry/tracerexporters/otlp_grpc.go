package tracerexporters

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// NewOTLPGrpc - Creates new OTLP exporter using GRPC protocol.
func NewOTLPGrpc(endpoint string, insecure bool, headers map[string]string) (trace.SpanExporter, error) {
	var exporter trace.SpanExporter
	var err error

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

	exporter, err = otlptracegrpc.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
