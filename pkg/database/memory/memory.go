package memory

import (
	"context"
	"sync"

	"github.com/hashicorp/go-memdb"
)

// Memory -
type Memory struct {
	sync.RWMutex

	DB *memdb.MemDB
}

// New -
func New(schema *memdb.DBSchema) (*Memory, error) {
	db, err := memdb.NewMemDB(schema)
	return &Memory{
		DB: db,
	}, err
}

// IsReady -
func (m *Memory) IsReady(ctx context.Context) (bool, error) {
	return true, nil
}

// GetEngineType -
func (m *Memory) GetEngineType() string {
	return "memory"
}

// Migrate -
func (m *Memory) Migrate(statements []string) (err error) {
	return nil
}

// Close -
func (m *Memory) Close() error {
	m.Lock()
	defer m.Unlock()
	m.DB = nil
	return nil
}
