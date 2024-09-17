package logexporters

import (
	"fmt"

	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
)

// ExporterFactory - Create log exporter according to given params
func ExporterFactory(name, endpoint string, insecure bool, urlpath string, headers map[string]string, protocol string) (*otlplogs.Exporter, error) {
	switch name {
	case "otlp", "otlp-http", "otlp-grpc":
		return NewOTLP(endpoint, insecure, urlpath, headers, protocol)
	default:
		return nil, fmt.Errorf("%s log exporter is unsupported", name)
	}
}
