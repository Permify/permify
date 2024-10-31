package internal

import "go.opentelemetry.io/otel"

var (
	tracer = otel.Tracer("permify")
	meter  = otel.Meter("permify")
)
