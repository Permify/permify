package tracerexporters

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// NewSigNoz = Creates new sigNoz tracer
func NewSigNoz(url string, insecure bool, headers map[string]string) (trace.SpanExporter, error) {
	var opts []otlptracegrpc.Option
	if insecure {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	} else {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}
	opts = append(opts, otlptracegrpc.WithEndpoint(url))
	return otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			opts...,
		),
	)
}
