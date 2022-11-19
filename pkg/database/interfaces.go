package database

import (
	"context"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Database -
type Database interface {
	// Migrate -
	Migrate(statements []string) error

	// IsReady returns true if the database is ready to be used.
	IsReady(ctx context.Context) (bool, error)

	// GetEngineType get the database type (e.g. postgres, memory, etc.).
	GetEngineType() string

	// Close the database connection.
	Close() error
}

// ITERATORS

// ITupleIterator abstract tuple iterator.
type ITupleIterator interface {
	HasNext() bool
	GetNext() *base.Tuple
}

// ISubjectIterator abstract subject iterator.
type ISubjectIterator interface {
	HasNext() bool
	GetNext() *base.Subject
}

// IEntityIterator abstract subject iterator.
type IEntityIterator interface {
	HasNext() bool
	GetNext() *base.Entity
}

// COLLECTIONS

// ITupleCollection abstract subject collection.
type ITupleCollection interface {
	Add(tuple *base.Tuple)
	CreateTupleIterator() ITupleIterator
	GetTuples() []*base.Tuple
	ToSubjectCollection() ISubjectCollection
}

// ISubjectCollection abstract subject collection.
type ISubjectCollection interface {
	Add(subject *base.Subject)
	CreateSubjectIterator() ISubjectIterator
	GetSubjects() []*base.Subject
}

// IEntityCollection abstract entity collection.
type IEntityCollection interface {
	Add(entity *base.Entity)
	CreateEntityIterator() IEntityIterator
	GetEntities() []*base.Entity
}
