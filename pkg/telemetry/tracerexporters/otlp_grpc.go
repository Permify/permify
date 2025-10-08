package tracerexporters // OTLP gRPC tracer exporter package
import (                // gRPC trace dependencies
	"context" // Context management

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc" // gRPC tracer
	"go.opentelemetry.io/otel/sdk/trace"                              // Trace SDK
	"google.golang.org/grpc/credentials"                              // TLS credentials
) // End imports

// NewOTLPGrpc - Creates new OTLP trace exporter using GRPC protocol.
func NewOTLPGrpc(
	endpoint string,
	insecure bool,
	headers map[string]string,
) (trace.SpanExporter, error) { // Create gRPC tracer exporter
	var spanExporter trace.SpanExporter // Span exporter instance
	var exportErr error                 // Error holder
	grpcOpts := []otlptracegrpc.Option{ // Configure gRPC options
		otlptracegrpc.WithEndpoint(endpoint), // Set endpoint
	} // Initial options
	if len(headers) > 0 { // Headers provided
		grpcOpts = append(grpcOpts, otlptracegrpc.WithHeaders(headers)) // Add headers
	} // Headers configured
	if insecure { // Insecure connection mode
		grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure()) // Disable TLS
	} else { // Secure connection mode
		grpcOpts = append(grpcOpts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))) // Enable TLS
	} // Security configured
	spanExporter, exportErr = otlptracegrpc.New(context.Background(), grpcOpts...) // Create exporter
	if exportErr != nil {                                                          // Creation failed
		return nil, exportErr // Return error
	} // Exporter created
	return spanExporter, nil // Return exporter
} // End NewOTLPGrpc
