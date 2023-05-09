package postgres

import (
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("storage.postgres")
