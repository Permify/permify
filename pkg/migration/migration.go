package migration

import (
	"fmt"
)

// Type -
type Type string

const (
	TABLE    Type = "table"
	FUNCTION Type = "function"
	TRIGGER  Type = "trigger"
)

// Migration -
type Migration struct {
	tables    map[string]string
	functions map[string]string
	triggers  map[string]string
}

// New -
func New() *Migration {
	return &Migration{
		tables:    make(map[string]string),
		functions: make(map[string]string),
		triggers:  make(map[string]string),
	}
}

// Register -
func (m *Migration) Register(typ Type, version string, sql string) (err error) {
	switch typ {
	case TABLE:
		if _, ok := m.tables[version]; ok {
			return fmt.Errorf("table version already exists: %s", version)
		}
		m.tables[version] = sql
		return
	case FUNCTION:
		if _, ok := m.functions[version]; ok {
			return fmt.Errorf("function version already exists: %s", version)
		}
		m.functions[version] = sql
	case TRIGGER:
		if _, ok := m.triggers[version]; ok {
			return fmt.Errorf("trigger version already exists: %s", version)
		}
		m.triggers[version] = sql
	default:
		return
	}
	return
}

// Tables -
func (m *Migration) Tables() map[string]string {
	return m.tables
}

// Functions -
func (m *Migration) Functions() map[string]string {
	return m.functions
}

// Triggers -
func (m *Migration) Triggers() map[string]string {
	return m.triggers
}
