package telemetry

import (
	"os"

	sdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

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
	)
}
