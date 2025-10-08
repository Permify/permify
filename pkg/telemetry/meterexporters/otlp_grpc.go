package meterexporters // OTLP gRPC meter exporter package
import (               // gRPC metric dependencies
	"context" // Context management
	// gRPC metrics
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc" // gRPC metrics
	"go.opentelemetry.io/otel/sdk/metric"                               // Metric SDK
	"google.golang.org/grpc/credentials"                                // TLS credentials
) // End imports
// NewOTLPGrpc - Creates new OTLP metric exporter using GRPC protocol for metrics collection.
func NewOTLPGrpc(
	endpoint string,
	insecure bool,
	headers map[string]string,
) (metric.Exporter, error) { // Create gRPC meter exporter
	grpcOpts := []otlpmetricgrpc.Option{ // Configure gRPC options
		otlpmetricgrpc.WithEndpoint(endpoint), // Set endpoint
		otlpmetricgrpc.WithHeaders(headers),   // Set headers
	} // Initial options
	if len(headers) > 0 { // Additional headers provided
		grpcOpts = append(grpcOpts, otlpmetricgrpc.WithHeaders(headers)) // Add extra headers
	} // Headers configured
	if insecure { // Insecure connection mode
		grpcOpts = append(grpcOpts, otlpmetricgrpc.WithInsecure()) // Disable TLS verification
	} else { // Secure connection mode
		grpcOpts = append(grpcOpts, otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))) // Enable TLS
	} // Security configured
	grpcExporter, exportErr := otlpmetricgrpc.New(context.Background(), grpcOpts...) // Create exporter instance
	if exportErr != nil {                                                            // Exporter creation failed
		return nil, exportErr // Return error
	} // Exporter created successfully
	return grpcExporter, nil // Return configured exporter
} // End NewOTLPGrpc
