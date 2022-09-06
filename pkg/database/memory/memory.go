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

// GetConnectionType -
func (m *Memory) GetConnectionType() string {
	return "memory"
}

// Close -
func (m *Memory) Close() {
	m.Lock()
	defer m.Unlock()
	m.DB = nil
}
