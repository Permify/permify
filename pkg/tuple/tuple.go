package tuple

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

const (
	ENTITY   = "%s:%s" // format string for entity in the form of "<type>:<id>"
	RELATION = "#%s"   // format string for relation in the form of "#<relation>"
)

const (
	ELLIPSIS = "..." // ellipsis string
)

const (
	USER = "user" // type string for user
)

const (
	SEPARATOR = "." // separator string used to concatenate entity and relation
)

// IsSubjectUser checks if the given subject is of type "user"
func IsSubjectUser(subject *base.Subject) bool {
	return subject.Type == USER
}

// AreSubjectsEqual checks if two subjects are equal
func AreSubjectsEqual(s1, s2 *base.Subject) bool {
	return s1.GetRelation() == s2.GetRelation() && s1.GetId() == s2.GetId() && s1.GetType() == s2.GetType()
}

// EntityAndRelationToString converts an EntityAndRelation object to string format
func EntityAndRelationToString(entityAndRelation *base.EntityAndRelation) string {
	return EntityToString(entityAndRelation.GetEntity()) + fmt.Sprintf(RELATION, entityAndRelation.GetRelation())
}

// EntityToString converts an Entity object to string format
func EntityToString(entity *base.Entity) string {
	return fmt.Sprintf(ENTITY, entity.GetType(), entity.GetId())
}

// SubjectToString converts a Subject object to string format
func SubjectToString(subject *base.Subject) string {
	if IsSubjectUser(subject) {
		return fmt.Sprintf(ENTITY, subject.GetType(), subject.GetId())
	}
	return fmt.Sprintf("%s"+RELATION, fmt.Sprintf(ENTITY, subject.GetType(), subject.GetId()), subject.GetRelation())
}

// IsEntityAndSubjectEquals checks if the entity and subject of a Tuple object are equal
func IsEntityAndSubjectEquals(t *base.Tuple) bool {
	return t.GetEntity().GetType() == t.GetSubject().GetType() && t.GetEntity().GetId() == t.GetSubject().GetId() && t.GetRelation() == t.GetSubject().GetRelation()
}

// ValidateSubjectType validates if the subject type and relation are present in the list of allowed relation types
func ValidateSubjectType(subject *base.Subject, relationTypes []string) (err error) {
	if len(relationTypes) == 0 {
		return errors.New(base.ErrorCode_ERROR_CODE_SUBJECT_TYPE_NOT_FOUND.String())
	}

	key := subject.GetType()
	if subject.GetRelation() != "" {
		if !IsSubjectUser(subject) { // if subject is not of type "user"
			if subject.GetRelation() != ELLIPSIS { // if subject relation is not an ellipsis
				key += "#" + subject.GetRelation() // append relation to key
			}
		}
	}

	if !slices.Contains(relationTypes, key) { // check if key is in relationTypes
		return errors.New(base.ErrorCode_ERROR_CODE_SUBJECT_TYPE_NOT_FOUND.String()) // return error if not found
	}
	return nil // return nil if validation succeeds
}

// SplitRelation splits a relation string by the separator "." and returns the result as a slice
func SplitRelation(relation string) (a []string) {
	s := strings.Split(relation, SEPARATOR) // split relation by the separator "."
	a = append(a, s...)
	if len(a) == 1 {
		a = append(a, "") // if there is only one element in the slice, add an empty string to the end
	}
	return
}

// IsRelationComputed checks if a relation is computed or not
func IsRelationComputed(relation string) bool {
	sp := strings.Split(relation, SEPARATOR)
	return len(sp) == 1
}

// IsSubjectValid checks if a subject is valid or not
func IsSubjectValid(subject *base.Subject) bool {
	if subject.GetType() == "" {
		return false
	}

	if subject.GetId() == "" {
		return false
	}

	if IsSubjectUser(subject) {
		return subject.GetRelation() == "" // relation should be empty for user subjects
	}
	return subject.GetRelation() != "" // relation should not be empty for non-user subjects
}

// Tuple parses a tuple string and returns a Tuple object
func Tuple(tuple string) (*base.Tuple, error) {
	s := strings.Split(strings.TrimSpace(tuple), "@") // split tuple string by "@"
	if len(s) != 2 {
		return nil, ErrInvalidTuple // return error if number of "@" is not equal to 2
	}
	ear, err := EAR(s[0]) // parse entity and relation from the first part of the tuple string
	if err != nil {
		return nil, err
	}
	sub, err := EAR(s[1]) // parse entity and relation from the second part of the tuple string
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

// EAR parses an EntityAndRelation string and returns an EntityAndRelation object
func EAR(ear string) (*base.EntityAndRelation, error) {
	s := strings.Split(strings.TrimSpace(ear), "#") // split EntityAndRelation string by "#"
	if len(s) == 1 {
		e, err := E(s[0]) // parse entity from the string
		if err != nil {
			return nil, err
		}
		return &base.EntityAndRelation{
			Entity:   e,
			Relation: "",
		}, nil
	} else if len(s) == 2 {
		e, err := E(s[0]) // parse entity from the string
		if err != nil {
			return nil, err
		}
		return &base.EntityAndRelation{
			Entity:   e,
			Relation: s[1],
		}, nil
	} else {
		return nil, ErrInvalidEntityAndRelation // return error if number of "#" is not equal to 1 or 2
	}
}

// E parses an Entity string and returns an Entity object
func E(e string) (*base.Entity, error) {
	s := strings.Split(strings.TrimSpace(e), ":") // split Entity string by ":"
	if len(s) != 2 {
		return nil, ErrInvalidEntity // return error if number of ":" is not equal to 2
	}
	return &base.Entity{
		Type: s[0],
		Id:   s[1],
	}, nil
}

// RelationReference - parses a relation reference string and returns a RelationReference object
func RelationReference(ref string) *base.RelationReference {
	sp := strings.Split(ref, "#")
	if len(sp) > 1 {
		return &base.RelationReference{
			Type:     sp[0],
			Relation: sp[1],
		}
	}
	return &base.RelationReference{
		Type:     sp[0],
		Relation: "",
	}
}

// AreRelationReferencesEqual checks if two relation references are equal or not
func AreRelationReferencesEqual(s1, s2 *base.RelationReference) bool {
	return s1.GetRelation() == s2.GetRelation() && s1.GetType() == s2.GetType()
}

// SetSubjectRelationToEllipsisIfNonUserAndNoRelation sets the relation of a subject to an ellipsis if the subject is not of type "user" and the relation is empty
func SetSubjectRelationToEllipsisIfNonUserAndNoRelation(subject *base.Subject) *base.Subject {
	if !IsSubjectUser(subject) && subject.GetRelation() == "" {
		subject.Relation = ELLIPSIS
	}
	return subject
}
