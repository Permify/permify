package tracerexporters

import (
	"testing"
)

func TestExporterFactory(t *testing.T) {
	tests := []struct {
		name       string
		exporter   string
		url        string
		protocol   string
		wantErr    bool
		errMessage string
	}{
		{name: "zipkin", exporter: "zipkin", url: "http://localhost:9411/api/v2/spans", protocol: "http", wantErr: false},
		{name: "jaeger", exporter: "jaeger", url: "http://localhost:14268/api/traces", protocol: "http", wantErr: false},
		{name: "otlp alias is accepted", exporter: "otlp", url: "localhost:4317", protocol: "http", wantErr: false},
		{name: "otlp-http alias is accepted", exporter: "otlp-http", url: "localhost:4317", protocol: "http", wantErr: false},
		{name: "otlp-grpc alias is accepted", exporter: "otlp-grpc", url: "localhost:4317", protocol: "grpc", wantErr: false},
		{name: "signoz", exporter: "signoz", url: "localhost:4317", protocol: "grpc", wantErr: false},
		{name: "unsupported exporter", exporter: "prometheus", url: "localhost:4317", protocol: "http", wantErr: true, errMessage: "prometheus tracer exporter is unsupported"},
		{name: "unsupported protocol", exporter: "otlp", url: "localhost:4317", protocol: "tcp", wantErr: true, errMessage: "unsupported protocol: tcp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, err := ExporterFactory(tt.exporter, tt.url, true, "", nil, tt.protocol)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errMessage != "" && err.Error() != tt.errMessage {
					t.Fatalf("expected error %q, got %q", tt.errMessage, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exp == nil {
				t.Fatalf("expected exporter, got nil")
			}
		})
	}
}

func TestNewOTLP(t *testing.T) {
	tests := []struct {
		name     string
		insecure bool
		urlpath  string
		headers  map[string]string
		protocol string
		wantErr  bool
	}{
		{name: "http insecure with options", insecure: true, urlpath: "/v1/traces", headers: map[string]string{"key": "value"}, protocol: "http"},
		{name: "http secure", insecure: false, protocol: "http"},
		{name: "grpc insecure with headers", insecure: true, headers: map[string]string{"key": "value"}, protocol: "grpc"},
		{name: "grpc secure", insecure: false, protocol: "grpc"},
		{name: "unsupported protocol", protocol: "udp", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, err := NewOTLP("localhost:4317", tt.insecure, tt.urlpath, tt.headers, tt.protocol)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exp == nil {
				t.Fatalf("expected exporter, got nil")
			}
		})
	}
}

func TestNewZipkin(t *testing.T) {
	exp, err := NewZipkin("http://localhost:9411/api/v2/spans")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp == nil {
		t.Fatalf("expected exporter, got nil")
	}
}

func TestNewJaegar(t *testing.T) {
	exp, err := NewJaegar("http://localhost:14268/api/traces")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp == nil {
		t.Fatalf("expected exporter, got nil")
	}
}

func TestNewSigNoz(t *testing.T) {
	tests := []struct {
		name     string
		insecure bool
		headers  map[string]string
	}{
		{name: "insecure with headers", insecure: true, headers: map[string]string{"key": "value"}},
		{name: "secure without headers", insecure: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, err := NewSigNoz("localhost:4317", tt.insecure, tt.headers)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exp == nil {
				t.Fatalf("expected exporter, got nil")
			}
		})
	}
}
