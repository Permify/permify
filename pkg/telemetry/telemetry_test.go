package telemetry

import (
	"context"
	"log/slog"
	"testing"

	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	"github.com/Permify/permify/pkg/telemetry/logexporters"
	"github.com/Permify/permify/pkg/telemetry/tracerexporters"
)

func TestNewResource(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		wantService string
	}{
		{name: "empty falls back to permify", serviceName: "", wantService: "permify"},
		{name: "whitespace falls back to permify", serviceName: "   ", wantService: "permify"},
		{name: "custom service name is honored", serviceName: "my-service", wantService: "my-service"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := newResource(tt.serviceName)
			if res == nil {
				t.Fatalf("expected resource, got nil")
			}

			var got string
			for _, attr := range res.Attributes() {
				if attr.Key == semconv.ServiceNameKey {
					got = attr.Value.AsString()
				}
			}
			if got != tt.wantService {
				t.Fatalf("expected service name %q, got %q", tt.wantService, got)
			}
		})
	}
}

func TestHandlerFactory(t *testing.T) {
	tests := []struct {
		name    string
		handler string
		headers map[string]string
		wantErr bool
	}{
		{name: "otlp", handler: "otlp", wantErr: false},
		{name: "gcp missing project errors", handler: "gcp", headers: map[string]string{}, wantErr: true},
		{name: "unsupported handler", handler: "prometheus", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := HandlerFactory(tt.handler, "localhost:4317", true, "", tt.headers, "http", slog.LevelInfo, "svc")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if h == nil {
				t.Fatalf("expected handler, got nil")
			}
		})
	}
}

func TestNewOTLPHandler(t *testing.T) {
	h, err := NewOTLPHandler("localhost:4317", true, "/v1/logs", map[string]string{"key": "value"}, "http", slog.LevelDebug, "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Fatalf("expected handler, got nil")
	}
}

func TestNewGCPHandlerMissingProject(t *testing.T) {
	if _, err := NewGCPHandler(map[string]string{}, slog.LevelInfo); err == nil {
		t.Fatalf("expected error when project id is missing")
	}
}

func TestNewLog(t *testing.T) {
	exporter, err := logexporters.NewOTLP("localhost:4317", true, "", nil, "http")
	if err != nil {
		t.Fatalf("failed to build exporter: %v", err)
	}
	lp := NewLog(exporter, "svc")
	if lp == nil {
		t.Fatalf("expected logger provider, got nil")
	}
	_ = lp.Shutdown(context.Background())
}

func TestNewTracer(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
	}{
		{name: "custom service name", serviceName: "svc"},
		{name: "empty service name", serviceName: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter, err := tracerexporters.NewOTLP("localhost:4317", true, "", nil, "http")
			if err != nil {
				t.Fatalf("failed to build exporter: %v", err)
			}
			shutdown := NewTracer(exporter, tt.serviceName)
			if shutdown == nil {
				t.Fatalf("expected shutdown func, got nil")
			}
			if err := shutdown(context.Background()); err != nil {
				t.Fatalf("unexpected shutdown error: %v", err)
			}
		})
	}
}

// NewMeter is intentionally not unit-tested here: it starts process-global
// gopsutil host instrumentation via host.Start and installs a global meter
// provider, which is environment-dependent and flaky under `go test -race`.
// Its constituent pieces (the OTLP meter exporter and the meter helpers below)
// are covered instead.

func TestNewNoopMeter(t *testing.T) {
	if NewNoopMeter() == nil {
		t.Fatalf("expected noop meter, got nil")
	}
}

func TestNewCounterAndHistogram(t *testing.T) {
	meter := NewNoopMeter()

	if NewCounter(meter, "test_counter", "a test counter") == nil {
		t.Fatalf("expected counter, got nil")
	}

	if NewHistogram(meter, "test_histogram", "ms", "a test histogram") == nil {
		t.Fatalf("expected histogram, got nil")
	}
}
