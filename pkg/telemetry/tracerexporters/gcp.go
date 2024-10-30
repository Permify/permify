package tracerexporters

import (
	"os"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// NewGCP creates a new Google Cloud tracer with optional headers for credentials and project ID.
func NewGCP(headers map[string]string) (sdktrace.SpanExporter, error) {
	if credentials, exists := headers["google-application-credentials"]; exists && credentials != "" {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentials); err != nil {
			return nil, err
		}
	}

	if projectID, exists := headers["google-cloud-project"]; exists && projectID != "" {
		exporter, err := texporter.New(texporter.WithProjectID(projectID))
		if err != nil {
			return nil, err
		}
		return exporter, nil
	}

	exporter, err := texporter.New()
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
