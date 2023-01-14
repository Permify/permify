package database

import (
	"context"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Database - Db interface
type Database interface {
	// Migrate -
	Migrate(statements []string) error

	// GetEngineType get the database type (e.g. postgres, memory, etc.).
	GetEngineType() string

	// Close the database connection.
	Close() error

	// IsReady - Check if database is ready
	IsReady(ctx context.Context) (bool, error)
}

// ITERATORS

// ITupleIterator - Abstract tuple iterator.
type ITupleIterator interface {
	HasNext() bool
	GetNext() *base.Tuple
}

// ISubjectIterator - Abstract subject iterator.
type ISubjectIterator interface {
	HasNext() bool
	GetNext() *base.Subject
}

// IEntityIterator - Abstract subject iterator.
type IEntityIterator interface {
	HasNext() bool
	GetNext() *base.Entity
}

// COLLECTIONS

// ITupleCollection - Abstract subject collection.
type ITupleCollection interface {
	Add(tuple *base.Tuple)
	CreateTupleIterator() ITupleIterator
	GetTuples() []*base.Tuple
	ToSubjectCollection() ISubjectCollection
}

// ISubjectCollection - Abstract subject collection.
type ISubjectCollection interface {
	Add(subject *base.Subject)
	CreateSubjectIterator() ISubjectIterator
	GetSubjects() []*base.Subject
}

// IEntityCollection - Abstract entity collection.
type IEntityCollection interface {
	Add(entity *base.Entity)
	CreateEntityIterator() IEntityIterator
	GetEntities() []*base.Entity
}
