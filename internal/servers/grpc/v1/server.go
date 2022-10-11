package v1

import (
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("grpc.servers")
