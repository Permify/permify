package ast

// Schema represents the parsed schema, which contains all the statements
// and extracted entity and relational references used by the schema. It
// is used as an intermediate representation before generating the
// corresponding metadata.
type Schema struct {
	// The list of statements in the schema
	Statements []Statement

	// references - Map of all relational references extracted from the schema
	references *References
}

func NewSchema() *Schema {
	return &Schema{
		Statements: []Statement{},
		references: NewReferences(),
	}
}

func (sch *Schema) GetReferences() *References {
	return sch.references
}

func (sch *Schema) SetReferences(refs *References) {
	sch.references = refs
}
