package postgres

const (
	RelationTuplesTable   = "relation_tuples"
	SchemaDefinitionTable = "schema_definitions"
	TransactionsTable     = "transactions"
	TenantsTable          = "tenants"
)

const (
	_defaultMaxTuplesPerWrite = 100
	_defaultMaxRetries        = 10
	_defaultWatchBufferSize   = 100
)
