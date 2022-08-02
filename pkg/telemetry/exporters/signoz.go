package exporters

import (
	`context`
	`go.opentelemetry.io/otel/exporters/otlp/otlptrace`
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	`go.opentelemetry.io/otel/sdk/trace`
	`google.golang.org/grpc/credentials`
)

// NewSigNoz =
func NewSigNoz(url string, insecure bool) (trace.SpanExporter, error) {
	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if !insecure {
		secureOption = otlptracegrpc.WithInsecure()
	}
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(url),
		),
	)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}
