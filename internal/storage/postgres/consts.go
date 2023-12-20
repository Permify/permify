package postgres

const (
	RelationTuplesTable   = "relation_tuples"
	AttributesTable       = "attributes"
	SchemaDefinitionTable = "schema_definitions"
	TransactionsTable     = "transactions"
	TenantsTable          = "tenants"
	BundlesTable          = "bundles"
)

const (
	_defaultMaxDataPerWrite = 100
	_defaultMaxRetries      = 10
	_defaultWatchBufferSize = 100
)
