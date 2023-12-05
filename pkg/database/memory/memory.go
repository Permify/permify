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

// Get - Gets a value from the in-memory database
func (m *Memory) Get(table, key string) (interface{}, error) {
	m.RLock()
	defer m.RUnlock()

	txn := m.DB.Txn(false) // Read-only txn
	raw, err := txn.First(table, "id", key)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

// New - Creates new database schema in memory
func New(schema *memdb.DBSchema) (*Memory, error) {
	db, err := memdb.NewMemDB(schema)
	return &Memory{
		DB: db,
	}, err
}

// Set - Sets a value in the in-memory database
func (m *Memory) Set(table, key string, value interface{}) error {
	m.Lock()
	defer m.Unlock()

	txn := m.DB.Txn(true) // Read-write txn
	err := txn.Insert(table, value)
	if err != nil {
		return err
	}
	txn.Commit()

	return nil
}

// Delete - Deletes a value from the in-memory database
func (m *Memory) Delete(table, key string) error {
	m.Lock()
	defer m.Unlock()

	txn := m.DB.Txn(true) // Read-write txn
	txn.Delete(table, key)
	txn.Commit()

	return nil
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
