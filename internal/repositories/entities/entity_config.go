package entities

import (
	"time"

	internalErrors "github.com/Permify/permify/internal/errors"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
	`github.com/Permify/permify/pkg/dsl/translator`
	"github.com/Permify/permify/pkg/errors"
)

// EntityConfig -
type EntityConfig struct {
	Entity           string    `json:"entity" bson:"entity"`
	SerializedConfig []byte    `json:"serialized_config" bson:"serialized_config"`
	Version          string    `json:"version" bson:"version"`
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
func (e EntityConfig) ToSchema() (schema.Schema, errors.Error) {
	var err error
	pr := parser.NewParser(string(e.SerializedConfig))
	parsed := pr.Parse()
	if pr.Error() != nil {
		return schema.Schema{}, pr.Error()
	}
	var s *translator.SchemaTranslator
	s, err = translator.NewSchemaTranslator(parsed)
	if err != nil {
		return schema.Schema{}, internalErrors.ConfigParserError
	}
	return s.Translate(), nil
}

// EntityConfigs -
type EntityConfigs []EntityConfig

// ToSchema -
func (e EntityConfigs) ToSchema() (schema.Schema, errors.Error) {
	var err error
	var configs string
	for _, c := range e {
		configs += string(c.SerializedConfig) + "\n"
	}
	pr := parser.NewParser(configs)
	parsed := pr.Parse()
	if pr.Error() != nil {
		return schema.Schema{}, internalErrors.ConfigParserError
	}
	var s *translator.SchemaTranslator
	s, err = translator.NewSchemaTranslator(parsed)
	if err != nil {
		return schema.Schema{}, internalErrors.ConfigParserError
	}
	return s.Translate(), nil
}
