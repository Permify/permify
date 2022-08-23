package tuple

import (
	"errors"
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/Permify/permify/pkg/helper"
)

const (
	ENTITY   = "%s:%s"
	RELATION = "#%s"
)

const (
	ELLIPSIS = "..."
)

const (
	USER = "user"
)

// EntityAndRelation -
type EntityAndRelation struct {
	Entity   Entity   `json:"entity"`
	Relation Relation `json:"relation"`
}

// String -
func (e EntityAndRelation) String() string {
	return e.Entity.String() + fmt.Sprintf(RELATION, e.Relation.String())
}

// Entity -
type Entity struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// String -
func (e Entity) String() string {
	return fmt.Sprintf(ENTITY, e.Type, e.ID)
}

// Validate -
func (e Entity) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&e,
		// object
		validation.Field(&e.Type, validation.Required),
		validation.Field(&e.ID, validation.Required),
	)
	return
}

// Subject -
type Subject struct {
	Type     string   `json:"type"`
	ID       string   `json:"id"`
	Relation Relation `json:"relation,omitempty"`
}

// Validate -
func (s Subject) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&s,
		// subject
		validation.Field(&s.Type, validation.Required),
		validation.Field(&s.ID, validation.Required),
		validation.Field(&s.Relation, validation.When(s.Type != USER, validation.Required).Else(validation.Empty)),
	)
	return
}

// Iterator -

type ISubjectIterator interface {
	HasNext() bool
	GetNext() *Subject
}

// SubjectIterator -
type SubjectIterator struct {
	index    int
	subjects []*Subject
}

// NewSubjectIterator -
func NewSubjectIterator(subjects []*Subject) *SubjectIterator {
	return &SubjectIterator{
		subjects: subjects,
	}
}

// HasNext -
func (u *SubjectIterator) HasNext() bool {
	if u.index < len(u.subjects) {
		return true
	}
	return false
}

// GetNext -
func (u *SubjectIterator) GetNext() *Subject {
	if u.HasNext() {
		t := u.subjects[u.index]
		u.index++
		return t
	}
	return nil
}

// String -
func (s Subject) String() string {
	if s.IsUser() {
		return fmt.Sprintf(ENTITY, s.Type, s.ID)
	}
	return fmt.Sprintf("%s"+RELATION, fmt.Sprintf(ENTITY, s.Type, s.ID), s.Relation)
}

// IsUser -
func (s Subject) IsUser() bool {
	if s.Type == USER {
		return true
	}
	return false
}

// IsValid -
func (s Subject) IsValid() bool {
	if s.Type == "" {
		return false
	}

	if s.ID == "" {
		return false
	}

	if s.IsUser() {
		if s.Type == USER && s.Relation == "" {
			return true
		}
	} else {
		if s.Relation != "" {
			return true
		}
	}

	return false
}

// Equals -
func (s Subject) Equals(v interface{}) bool {
	uv, ok := v.(Subject)
	if !ok {
		return false
	}
	if s.IsUser() {
		return s.ID == uv.ID
	}
	return uv.Relation == s.Relation && uv.ID == s.ID && uv.Type == s.Type
}

// ValidateSubjectType -
func (s Subject) ValidateSubjectType(relationTypes []string) (err error) {
	key := s.Type
	if s.Relation.String() != "" {
		if !s.IsUser() {
			if s.Relation.String() == ELLIPSIS {
				key += "#..."
			} else {
				key += "#" + s.Relation.String()
			}
		}
	}
	if !helper.InArray(key, relationTypes) {
		return NotFoundInSpecifiedRelationTypes
	}
	return nil
}

// Tuple -
type Tuple struct {
	Entity   Entity  `json:"entity"`
	Relation string  `json:"relation"`
	Subject  Subject `json:"subject"`
}

// String -
func (r Tuple) String() string {
	object := fmt.Sprintf(ENTITY, r.Entity.Type, r.Entity.ID)
	relation := fmt.Sprintf(RELATION, r.Relation)
	return object + relation + "@" + r.Subject.String()
}

// ConvertSubject -
func ConvertSubject(v string) Subject {
	parts := strings.Split(v, "#")
	if len(parts) != 2 {
		return Subject{
			Type: USER,
			ID:   v,
		}
	}

	innerParts := strings.Split(parts[0], ":")
	if len(innerParts) != 2 {
		return Subject{
			Type: USER,
			ID:   v,
		}
	}

	return Subject{
		Type:     innerParts[0],
		ID:       innerParts[1],
		Relation: Relation(parts[1]),
	}
}

// ConvertEntity -
func ConvertEntity(v string) (Entity, error) {
	obj := strings.Split(v, ":")
	if len(obj) < 2 {
		return Entity{}, errors.New("input is not suitable for the object")
	}
	return Entity{
		Type: obj[0],
		ID:   obj[1],
	}, nil
}

const (
	SEPARATOR = "."
)

// Relation -
type Relation string

// String -
func (r Relation) String() string {
	return string(r)
}

// IsComputed -
func (r Relation) IsComputed() bool {
	sp := strings.Split(r.String(), SEPARATOR)
	if len(sp) == 1 {
		return true
	}
	return false
}

// Split -
func (r Relation) Split() (a []Relation) {
	s := strings.Split(r.String(), SEPARATOR)
	for _, b := range s {
		a = append(a, Relation(b))
	}
	if len(a) == 1 {
		a = append(a, "")
	}
	return
}
