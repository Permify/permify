package postgres

const (
	RelationTuplesTable   = "relation_tuples"
	SchemaDefinitionTable = "schema_definitions"
	TransactionsTable     = "transactions"
)

const (
	_defaultMaxTuplesPerWrite = 100
	_defaultMaxRetries        = 10
)
