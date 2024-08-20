package memory

import (
	"context"
	"sync"

	"github.com/hashicorp/go-memdb"
)

// Memory - Structure for in memory db
type Memory struct {
	sync.RWMutex
	rid uint64
	aid uint64

	DB *memdb.MemDB
}

// New - Creates new database schema in memory
func New(schema *memdb.DBSchema) (*Memory, error) {
	db, err := memdb.NewMemDB(schema)
	return &Memory{
		DB: db,
	}, err
}

func (m *Memory) RelationTupleID() (id uint64) {
	m.Lock()
	defer m.Unlock()
	if m.rid == 0 {
		m.rid++
	}
	id = m.rid
	m.rid++
	return
}

func (m *Memory) AttributeID() (id uint64) {
	m.Lock()
	defer m.Unlock()
	if m.aid == 0 {
		m.aid++
	}
	id = m.aid
	m.aid++
	return
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
