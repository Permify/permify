package storage

import (
	"time"

	"google.golang.org/protobuf/types/known/anypb"

	"google.golang.org/protobuf/types/known/timestamppb"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// RelationTuple - Structure for Relational Tuple
type RelationTuple struct {
	ID              uint64
	TenantID        string
	EntityType      string
	EntityID        string
	Relation        string
	SubjectType     string
	SubjectID       string
	SubjectRelation string
}

// ToTuple - Convert database relation tuple to base relation tuple
func (r RelationTuple) ToTuple() *base.Tuple {
	return &base.Tuple{
		Entity: &base.Entity{
			Type: r.EntityType,
			Id:   r.EntityID,
		},
		Relation: r.Relation,
		Subject: &base.Subject{
			Type:     r.SubjectType,
			Id:       r.SubjectID,
			Relation: r.SubjectRelation,
		},
	}
}

type Attribute struct {
	ID         uint64
	TenantID   string
	EntityType string
	EntityID   string
	Attribute  string
	Value      *anypb.Any
}

func (r Attribute) ToAttribute() *base.Attribute {
	return &base.Attribute{
		Entity: &base.Entity{
			Type: r.EntityType,
			Id:   r.EntityID,
		},
		Attribute: r.Attribute,
		Value:     r.Value,
	}
}

// SchemaDefinition - Structure for Schema Definition
type SchemaDefinition struct {
	TenantID             string
	Name                 string
	SerializedDefinition []byte
	Version              string
}

// Serialized - get schema serialized definition
func (e SchemaDefinition) Serialized() string {
	return string(e.SerializedDefinition)
}

// Tenant - Structure for tenant
type Tenant struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

// ToTenant - Convert database tenant to base tenant
func (r Tenant) ToTenant() *base.Tenant {
	return &base.Tenant{
		Id:        r.ID,
		Name:      r.Name,
		CreatedAt: timestamppb.New(r.CreatedAt),
	}
}

// Bundle - Structure for Bundle
type Bundle struct {
	Name       string
	DataBundle *base.DataBundle
	TenantID   string
}
