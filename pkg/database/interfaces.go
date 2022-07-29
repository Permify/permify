package database

// Database -
type Database interface {
	GetConnectionType() string
	Close()
}
