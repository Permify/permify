package telemetry

import (
	"os"
	"runtime"

	"go.opentelemetry.io/otel/attribute"

	sdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	"github.com/Permify/permify/internal"
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
		attribute.String("id", internal.Identifier),
		attribute.String("project.id", internal.Identifier),
		attribute.String("version", internal.Version),
		attribute.String("host_name", hostName),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
	)
}
