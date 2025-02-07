package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/Permify/sloggcp"
	"github.com/agoda-com/opentelemetry-go/otelslog"

	"go.opentelemetry.io/otel/attribute"

	sdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/pkg/telemetry/logexporters"
)

// HandlerFactory - Create log handler according to given params
func HandlerFactory(name, endpoint string, insecure bool, urlpath string, headers map[string]string, protocol string, level slog.Leveler) (slog.Handler, error) {
	switch name {
	case "otlp", "otlp-http", "otlp-grpc":
		return NewOTLPHandler(endpoint, insecure, urlpath, headers, protocol, level.Level())
	case "gcp":
		return NewGCPHandler(headers, level)
	default:
		return nil, fmt.Errorf("%s log handler is unsupported", name)
	}
}

func NewOTLPHandler(endpoint string, insecure bool, urlPath string, headers map[string]string, protocol string, level slog.Leveler) (slog.Handler, error) {
	// Set up the OTLP exporter based on the protocol
	exporter, err := logexporters.ExporterFactory("otlp", endpoint, insecure, urlPath, headers, protocol)
	if err != nil {
		return nil, errors.New("failed to create OTLP exporter")
	}

	// Initialize the OpenTelemetry handler with the exporter
	lp := NewLog(exporter)
	otelHandler := otelslog.NewOtelHandler(lp, &otelslog.HandlerOptions{
		Level: level,
	})

	// Shut down the exporter when needed
	return otelHandler, nil
}

func NewGCPHandler(headers map[string]string, level slog.Leveler) (slog.Handler, error) {
	// Retrieve Google Cloud credentials from headers
	creds := headers["google-application-credentials"]
	projectId := headers["google-cloud-project"]

	if projectId == "" {
		return nil, errors.New("missing GOOGLE_CLOUD_PROJECT in headers")
	}

	// Set credentials for Google Cloud access
	if creds != "" {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", creds); err != nil {
			return nil, err
		}
	}

	// Initialize GCP-specific log handler
	logName := internal.Identifier
	gcpHandler := sloggcp.NewGoogleCloudSlogHandler(context.Background(), projectId, logName, &slog.HandlerOptions{
		Level: level,
	})
	return gcpHandler, nil
}

// NewLog - Creates new log
func NewLog(exporter sdk.LogRecordExporter) *sdk.LoggerProvider {
	// Create a logger provider with the exporter and resource
	lp := sdk.NewLoggerProvider(
		sdk.WithBatcher(exporter),
		sdk.WithResource(newResource()),
	)

	// Return the logger provider
	return lp
}

func newResource() *resource.Resource {
	hostName, _ := os.Hostname()
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("permify"),
		semconv.HostNameKey.String(hostName),
		attribute.String("id", internal.Identifier),
		attribute.String("project.id", internal.Identifier),
		attribute.String("version", internal.Version),
		attribute.String("host_name", hostName),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
	)
}
