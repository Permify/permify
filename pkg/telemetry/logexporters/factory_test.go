package logexporters

import (
	"testing"
)

func TestExporterFactory(t *testing.T) {
	tests := []struct {
		name       string
		exporter   string
		protocol   string
		wantErr    bool
		errMessage string
	}{
		{name: "otlp alias is accepted", exporter: "otlp", protocol: "http", wantErr: false},
		{name: "otlp-http alias is accepted", exporter: "otlp-http", protocol: "http", wantErr: false},
		{name: "otlp-grpc alias is accepted", exporter: "otlp-grpc", protocol: "grpc", wantErr: false},
		{name: "unsupported exporter", exporter: "prometheus", protocol: "http", wantErr: true, errMessage: "prometheus log exporter is unsupported"},
		{name: "unsupported protocol", exporter: "otlp", protocol: "tcp", wantErr: true, errMessage: "unsupported protocol: tcp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, err := ExporterFactory(tt.exporter, "localhost:4317", true, "", nil, tt.protocol)
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
		{name: "http insecure with options", insecure: true, urlpath: "/v1/logs", headers: map[string]string{"key": "value"}, protocol: "http"},
		{name: "http secure", insecure: false, protocol: "http"},
		{name: "grpc insecure", insecure: true, protocol: "grpc"},
		{name: "grpc secure with headers", insecure: false, headers: map[string]string{"key": "value"}, protocol: "grpc"},
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
