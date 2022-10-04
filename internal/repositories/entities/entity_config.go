package entities

import (
	"time"

	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/dsl/translator"
	"github.com/Permify/permify/pkg/errors"
)

// EntityConfig -
type EntityConfig struct {
	Entity           string    `json:"entity"`
	SerializedConfig []byte    `json:"serialized_config"`
	Version          string    `json:"version"`
	CommitTime       time.Time `json:"commit_time"`
}

// Table -
func (EntityConfig) Table() string {
	return "entity_config"
}

// ToSchema -
func (e EntityConfig) ToSchema() (schema.Schema, errors.Error) {
	var err error
	pr, pErr := parser.NewParser(string(e.SerializedConfig)).Parse()
	if pErr != nil {
		return schema.Schema{}, pErr
	}
	var s *translator.SchemaTranslator
	s, err = translator.NewSchemaTranslator(pr)
	if err != nil {
		return schema.Schema{}, errors.ValidationError.AddParam("schema", err.Error())
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
	pr, pErr := parser.NewParser(configs).Parse()
	if pErr != nil {
		return schema.Schema{}, pErr
	}
	var s *translator.SchemaTranslator
	s, err = translator.NewSchemaTranslator(pr)
	if err != nil {
		return schema.Schema{}, errors.ValidationError.AddParam("schema", err.Error())
	}
	return s.Translate(), nil
}
