package memory

import (
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("storage.memory")
