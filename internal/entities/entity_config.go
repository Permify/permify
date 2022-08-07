package entities

import (
	"time"
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

// EntityConfigs -
type EntityConfigs []EntityConfig
