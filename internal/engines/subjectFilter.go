package engines

import (
	"context"
	"errors"

	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type SubjectFilterEngine struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// relationshipReader is responsible for reading relationship information
	relationshipReader storage.RelationshipReader
	concurrencyLimit   int
}

func NewSubjectFilterEngine(sr storage.SchemaReader, rr storage.RelationshipReader) *SubjectFilterEngine {
	// Initialize a CheckEngine with default concurrency limit and provided parameters
	engine := &SubjectFilterEngine{
		schemaReader:       sr,
		relationshipReader: rr,
		concurrencyLimit:   100,
	}

	return engine
}

func (engine *SubjectFilterEngine) SubjectFilter(
	ctx context.Context,
	request *base.PermissionSubjectFilterRequest,
	stream *Stream,
) (err error) {
	return errors.New("implementing")
}
