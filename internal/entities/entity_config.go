package entities

import (
	"time"
)

// EntityConfig -
type EntityConfig struct {
	Entity           string    `json:"entity"`
	SerializedConfig []byte    `json:"serialized_config"`
	CommitTime       time.Time `json:"commit_time"`
}

// Table -
func (EntityConfig) Table() string {
	return "entity_config"
}
