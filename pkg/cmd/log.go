package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/Permify/sloggcp"
	"github.com/agoda-com/opentelemetry-go/otelslog"

	"github.com/Permify/permify/pkg/telemetry"
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
	lp := telemetry.NewLog(exporter)
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
	logName := "permify"
	gcpHandler := sloggcp.NewGoogleCloudSlogHandler(context.Background(), projectId, logName, &slog.HandlerOptions{
		Level: level,
	})
	return gcpHandler, nil
}
