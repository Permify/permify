package database

// ConnectionType -
type ConnectionType string

const (
	POSTGRES ConnectionType = "postgres"
	MONGO    ConnectionType = "mongo"
)

// String -
func (c ConnectionType) String() string {
	return string(c)
}
