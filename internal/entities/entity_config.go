package entities

import (
	"time"

	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
)

// EntityConfig -
type EntityConfig struct {
	Entity           string    `json:"entity" bson:"entity"`
	SerializedConfig []byte    `json:"serialized_config" bson:"serialized_config"`
	CommitTime       time.Time `json:"commit_time" bson:"commit_time"`
}

// Table -
func (EntityConfig) Table() string {
	return "entity_config"
}

// Collection -
func (EntityConfig) Collection() string {
	return "entity_config"
}

// ToSchema -
func (e EntityConfig) ToSchema() (sch schema.Schema, err error) {
	pr := parser.NewParser(string(e.SerializedConfig))
	parsed := pr.Parse()
	if pr.Error() != nil {
		return schema.Schema{}, pr.Error()
	}
	var s *parser.SchemaTranslator
	s, err = parser.NewSchemaTranslator(parsed)
	if err != nil {
		return schema.Schema{}, err
	}
	return s.Translate(), nil
}

// EntityConfigs -
type EntityConfigs []EntityConfig

// ToSchema -
func (e EntityConfigs) ToSchema() (sch schema.Schema, err error) {
	var configs string
	for _, c := range e {
		configs += string(c.SerializedConfig) + "\n"
	}
	pr := parser.NewParser(configs)
	parsed := pr.Parse()
	if pr.Error() != nil {
		return schema.Schema{}, pr.Error()
	}
	var s *parser.SchemaTranslator
	s, err = parser.NewSchemaTranslator(parsed)
	if err != nil {
		return schema.Schema{}, err
	}
	return s.Translate(), nil
}
