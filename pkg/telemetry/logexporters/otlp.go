package logexporters // OTLP log exporter package
import (             // OTLP dependencies
	"context" // Context management
	"fmt"     // Error formatting utilities

	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"              // OTLP logs
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogsgrpc" // gRPC client
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogshttp" // HTTP client
) // End imports

// NewOTLP - Creates new OTLP exporter based on protocol (HTTP or gRPC).
func NewOTLP(endpoint string, insecure bool, urlpath string, headers map[string]string, protocol string) (*otlplogs.Exporter, error) { // Create OTLP log exporter
	switch protocol { // Determine protocol type
	case "http": // HTTP protocol configuration
		httpOptions := []otlplogshttp.Option{ // Configure HTTP options
			otlplogshttp.WithCompression(otlplogshttp.GzipCompression), // Enable gzip compression
			otlplogshttp.WithEndpoint(endpoint),                        // Set endpoint URL
		} // Initial HTTP options
		if urlpath != "" { // Custom URL path provided
			httpOptions = append(httpOptions, otlplogshttp.WithURLPath(urlpath)) // Add URL path
		} // URL path configured
		if insecure { // Insecure connection requested
			httpOptions = append(httpOptions, otlplogshttp.WithInsecure()) // Disable TLS
		} // HTTP security configured
		if len(headers) > 0 { // Headers provided
			httpOptions = append(httpOptions, otlplogshttp.WithHeaders(headers)) // Add headers
		} // Headers configured
		httpExporter, httpErr := otlplogs.NewExporter(context.Background(), otlplogs.WithClient(
			otlplogshttp.NewClient(httpOptions...), // Create HTTP client
		)) // Create HTTP exporter
		if httpErr != nil { // HTTP export creation failed
			return nil, httpErr // Return error
		} // HTTP exporter created successfully
		return httpExporter, nil // Return HTTP exporter
	case "grpc": // gRPC protocol configuration
		grpcOptions := []otlplogsgrpc.Option{ // Configure gRPC options
			otlplogsgrpc.WithEndpoint(endpoint), // Set gRPC endpoint
		} // Initial gRPC options
		if insecure { // Insecure connection requested
			grpcOptions = append(grpcOptions, otlplogsgrpc.WithInsecure()) // Disable TLS for gRPC
		} // gRPC security configured
		if len(headers) > 0 { // Headers provided
			grpcOptions = append(grpcOptions, otlplogsgrpc.WithHeaders(headers)) // Add headers
		} // Headers configured
		grpcExporter, grpcErr := otlplogs.NewExporter(context.Background(), otlplogs.WithClient(
			otlplogsgrpc.NewClient(grpcOptions...), // Create gRPC client
		)) // Create gRPC exporter
		if grpcErr != nil { // gRPC export creation failed
			return nil, grpcErr // Return error
		} // gRPC exporter created successfully
		return grpcExporter, nil // Return gRPC exporter
	default: // Unsupported protocol
		return nil, fmt.Errorf("unsupported protocol: %s", protocol) // Return error
	} // End protocol switch
} // End NewOTLP
