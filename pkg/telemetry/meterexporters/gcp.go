package meterexporters

import (
	"os"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

// NewGCP creates a new Google Cloud metric exporter with optional headers for credentials and project ID.
func NewGCP(headers map[string]string) (metric.Exporter, error) {
	if credentials, exists := headers["google-application-credentials"]; exists && credentials != "" {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentials); err != nil {
			return nil, err
		}
	}

	if projectID, exists := headers["google-cloud-project"]; exists && projectID != "" {
		exporter, err := mexporter.New(mexporter.WithProjectID(projectID))
		if err != nil {
			return nil, err
		}
		return exporter, nil
	}

	exporter, err := mexporter.New()
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
