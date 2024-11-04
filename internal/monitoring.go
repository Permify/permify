package internal

import "go.opentelemetry.io/otel"

var (
	Tracer = otel.Tracer("permify")
	Meter  = otel.Meter("permify")
)
