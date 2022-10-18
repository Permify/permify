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
			return errors.New(base.ErrorCode_subject_relation_must_be_empty.String())
		}
	} else {
		if subject.GetRelation() == "" {
			return errors.New(base.ErrorCode_subject_relation_cannot_be_empty.String())
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
		return errors.New(base.ErrorCode_subject_type_not_found.String())
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
		return errors.New(base.ErrorCode_subject_type_not_found.String())
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
