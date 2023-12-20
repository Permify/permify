package memory

import (
	"context"
	"sync"

	"github.com/hashicorp/go-memdb"
)

// Memory - Structure for in memory db
type Memory struct {
	sync.RWMutex

	DB *memdb.MemDB
}

// New - Creates new database schema in memory
func New(schema *memdb.DBSchema) (*Memory, error) {
	db, err := memdb.NewMemDB(schema)
	return &Memory{
		DB: db,
	}, err
}

// GetEngineType - Gets engine type, returns as string
func (m *Memory) GetEngineType() string {
	return "memory"
}

// Close - Closing the in memory instance
func (m *Memory) Close() error {
	m.Lock()
	defer m.Unlock()
	m.DB = nil
	return nil
}

// IsReady - Check if database is ready
func (m *Memory) IsReady(_ context.Context) (bool, error) {
	return true, nil
}
