package tracerexporters // OTLP tracer exporter package
import (                // OTLP trace dependencies
	"context" // Context management
	"fmt"     // Error formatting

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc" // gRPC tracer
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp" // HTTP tracer
	"go.opentelemetry.io/otel/sdk/trace"                              // Trace SDK
	"google.golang.org/grpc/credentials"                              // TLS credentials
) // End imports
// NewOTLP - Creates new OTLP trace exporter based on protocol (HTTP or gRPC).
func NewOTLP(endpoint string, insecure bool, urlpath string, headers map[string]string, protocol string) (trace.SpanExporter, error) { // Create OTLP tracer
	switch protocol { // Select protocol
	case "http": // HTTP configuration
		httpOpts := []otlptracehttp.Option{ // HTTP options
			otlptracehttp.WithEndpoint(endpoint), // Endpoint
		} // Base options
		if len(headers) > 0 { // Headers provided
			httpOpts = append(httpOpts, otlptracehttp.WithHeaders(headers)) // Add headers
		} // Headers configured
		if urlpath != "" { // URL path provided
			httpOpts = append(httpOpts, otlptracehttp.WithURLPath(urlpath)) // Add path
		} // Path configured
		if insecure { // Insecure mode
			httpOpts = append(httpOpts, otlptracehttp.WithInsecure()) // Disable TLS
		} // Security configured
		httpExp, httpErr := otlptracehttp.New(context.Background(), httpOpts...) // Create HTTP exporter
		if httpErr != nil {                                                      // Creation failed
			return nil, httpErr // Return error
		} // Exporter created
		return httpExp, nil // Return exporter
	case "grpc": // gRPC configuration
		grpcOpts := []otlptracegrpc.Option{ // gRPC options
			otlptracegrpc.WithEndpoint(endpoint), // Endpoint
		} // Base options
		if len(headers) > 0 { // Headers provided
			grpcOpts = append(grpcOpts, otlptracegrpc.WithHeaders(headers)) // Add headers
		} // Headers configured
		if insecure { // Insecure mode
			grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure()) // Disable TLS
		} else { // Secure mode
			grpcOpts = append(grpcOpts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))) // Enable TLS
		} // Security configured
		grpcExp, grpcErr := otlptracegrpc.New(context.Background(), grpcOpts...) // Create gRPC exporter
		if grpcErr != nil {                                                      // Creation failed
			return nil, grpcErr // Return error
		} // Exporter created
		return grpcExp, nil // Return exporter
	default: // Unknown protocol
		return nil, fmt.Errorf("unsupported protocol: %s", protocol) // Error
	} // End switch
} // End NewOTLP
