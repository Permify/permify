package meterexporters // OTLP meter exporter package
import (               // OTLP metric dependencies
	"context" // Context management
	"fmt"     // Error formatting

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc" // gRPC metrics
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp" // HTTP metrics
	"go.opentelemetry.io/otel/sdk/metric"                               // Metric SDK
	"google.golang.org/grpc/credentials"                                // TLS credentials
) // End imports

// NewOTLP - Creates new OTLP metric exporter based on protocol (HTTP or gRPC).
func NewOTLP(
	endpoint string,
	insecure bool,
	urlpath string,
	headers map[string]string,
	protocol string,
) (metric.Exporter, error) { // Create OTLP meter
	switch protocol { // Select protocol
	case "http": // HTTP configuration
		httpOpts := []otlpmetrichttp.Option{ // HTTP options
			otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression), // Gzip
			otlpmetrichttp.WithEndpoint(endpoint),                          // Endpoint
		} // Base options
		if len(headers) > 0 { // Headers provided
			httpOpts = append(httpOpts, otlpmetrichttp.WithHeaders(headers)) // Add headers
		} // Headers configured
		if urlpath != "" { // URL path provided
			httpOpts = append(httpOpts, otlpmetrichttp.WithURLPath(urlpath)) // Add path
		} // Path configured
		if insecure { // Insecure mode
			httpOpts = append(httpOpts, otlpmetrichttp.WithInsecure()) // Disable TLS
		} // Security configured
		httpExp, httpErr := otlpmetrichttp.New(context.Background(), httpOpts...) // Create HTTP exporter
		if httpErr != nil {                                                       // Creation failed
			return nil, httpErr // Return error
		} // Exporter created
		return httpExp, nil // Return exporter
	case "grpc": // gRPC configuration
		grpcOpts := []otlpmetricgrpc.Option{ // gRPC options
			otlpmetricgrpc.WithEndpoint(endpoint), // Endpoint
			otlpmetricgrpc.WithHeaders(headers),   // Headers
		} // Base options
		if insecure { // Insecure mode
			grpcOpts = append(grpcOpts, otlpmetricgrpc.WithInsecure()) // Disable TLS
		} else { // Secure mode
			grpcOpts = append(grpcOpts, otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))) // Enable TLS
		} // Security configured
		grpcExp, grpcErr := otlpmetricgrpc.New(context.Background(), grpcOpts...) // Create gRPC exporter
		if grpcErr != nil {                                                       // Creation failed
			return nil, grpcErr // Return error
		} // Exporter created
		return grpcExp, nil // Return exporter
	default: // Unknown protocol
		return nil, fmt.Errorf("unsupported protocol: %s", protocol) // Error
	} // End switch
} // End NewOTLP
