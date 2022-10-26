package tuple

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Permify/permify/pkg/helper"
	base "github.com/Permify/permify/pkg/pb/base/v1"
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

const (
	SEPARATOR = "."
)

// IsSubjectUser -
func IsSubjectUser(subject *base.Subject) bool {
	if subject.Type == USER {
		return true
	}
	return false
}

// ValidateSubject -
func ValidateSubject(subject *base.Subject) error {
	if subject.Type == USER {
		if subject.GetRelation() != "" {
			return errors.New(base.ErrorCode_ERROR_CODE_SUBJECT_RELATION_MUST_BE_EMPTY.String())
		}
	} else {
		if subject.GetRelation() == "" {
			return errors.New(base.ErrorCode_ERROR_CODE_SUBJECT_RELATION_CANNOT_BE_EMPTY.String())
		}
	}
	return nil
}

// AreSubjectsEqual -
func AreSubjectsEqual(s1 *base.Subject, s2 *base.Subject) bool {
	return s1.GetRelation() == s2.GetRelation() && s1.GetId() == s2.GetId() && s1.GetType() == s2.GetType()
}

// EntityAndRelationToString -
func EntityAndRelationToString(entityAndRelation *base.EntityAndRelation) string {
	return EntityToString(entityAndRelation.GetEntity()) + fmt.Sprintf(RELATION, entityAndRelation.GetRelation())
}

// EntityToString -
func EntityToString(entity *base.Entity) string {
	return fmt.Sprintf(ENTITY, entity.GetType(), entity.GetId())
}

// SubjectToString -
func SubjectToString(subject *base.Subject) string {
	if IsSubjectUser(subject) {
		return fmt.Sprintf(ENTITY, subject.GetType(), subject.GetId())
	}
	return fmt.Sprintf("%s"+RELATION, fmt.Sprintf(ENTITY, subject.GetType(), subject.GetId()), subject.GetRelation())
}

// IsEntityAndSubjectEquals -
func IsEntityAndSubjectEquals(t *base.Tuple) bool {
	return t.GetEntity().GetType() == t.GetSubject().GetType() && t.GetEntity().GetType() == t.GetSubject().GetType() && t.GetRelation() == t.GetSubject().GetRelation()
}

// ValidateSubjectType -
func ValidateSubjectType(subject *base.Subject, relationTypes []string) (err error) {
	if len(relationTypes) == 0 {
		return errors.New(base.ErrorCode_ERROR_CODE_SUBJECT_TYPE_NOT_FOUND.String())
	}

	key := subject.GetType()
	if subject.GetRelation() != "" {
		if !IsSubjectUser(subject) {
			if subject.GetRelation() != ELLIPSIS {
				key += "#" + subject.GetRelation()
			}
		}
	}

	if !helper.InArray(key, relationTypes) {
		return errors.New(base.ErrorCode_ERROR_CODE_SUBJECT_TYPE_NOT_FOUND.String())
	}
	return nil
}

// SplitRelation -
func SplitRelation(relation string) (a []string) {
	s := strings.Split(relation, SEPARATOR)
	for _, b := range s {
		a = append(a, b)
	}
	if len(a) == 1 {
		a = append(a, "")
	}
	return
}

// IsRelationComputed -
func IsRelationComputed(relation string) bool {
	sp := strings.Split(relation, SEPARATOR)
	if len(sp) == 1 {
		return true
	}
	return false
}

// IsSubjectValid -
func IsSubjectValid(subject *base.Subject) bool {
	if subject.GetType() == "" {
		return false
	}

	if subject.GetId() == "" {
		return false
	}

	if IsSubjectUser(subject) {
		return subject.GetRelation() == ""
	} else {
		return subject.GetRelation() != ""
	}
}

// Tuple -
func Tuple(tuple string) (*base.Tuple, error) {
	s := strings.Split(strings.TrimSpace(tuple), "@")
	if len(s) != 2 {
		return nil, ErrInvalidTuple
	}
	ear, err := EAR(s[0])
	if err != nil {
		return nil, err
	}
	sub, err := EAR(s[1])
	if err != nil {
		return nil, err
	}
	return &base.Tuple{
		Entity:   ear.Entity,
		Relation: ear.Relation,
		Subject: &base.Subject{
			Type:     sub.Entity.Type,
			Id:       sub.Entity.Id,
			Relation: sub.Relation,
		},
	}, nil
}

// EAR EntityAndRelation
func EAR(ear string) (*base.EntityAndRelation, error) {
	s := strings.Split(strings.TrimSpace(ear), "#")
	if len(s) == 1 {
		e, err := E(s[0])
		if err != nil {
			return nil, err
		}
		return &base.EntityAndRelation{
			Entity:   e,
			Relation: "",
		}, nil
	} else if len(s) == 2 {
		e, err := E(s[0])
		if err != nil {
			return nil, err
		}
		return &base.EntityAndRelation{
			Entity:   e,
			Relation: s[1],
		}, nil
	} else {
		return nil, ErrInvalidEntityAndRelation
	}
}

// E New Entity from string
func E(e string) (*base.Entity, error) {
	s := strings.Split(strings.TrimSpace(e), ":")
	if len(s) != 2 {
		return nil, ErrInvalidEntity
	}
	return &base.Entity{
		Type: s[0],
		Id:   s[1],
	}, nil
}

type Query struct {
	Subject *base.Subject
	Entity  *base.Entity
	Action  string
}

// NewQueryFromString sample query: can user:1 push repository:1
func NewQueryFromString(query string) (*Query, error) {
	q := strings.Split(strings.TrimSpace(query), " ")
	if len(q) != 4 {
		return nil, ErrInvalidQuery
	}
	subject, err := EAR(q[1])
	if err != nil {
		return nil, err
	}
	entity, err := E(q[3])
	if err != nil {
		return nil, err
	}
	return &Query{
		Entity: entity,
		Subject: &base.Subject{
			Type:     subject.Entity.Type,
			Id:       subject.Entity.Id,
			Relation: subject.Relation,
		},
		Action: q[2],
	}, nil
}
