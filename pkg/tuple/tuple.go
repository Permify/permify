package tuple

import (
	"errors"
	"fmt"
	"strings"
)

const (
	OBJECT   = "%s:%s"
	RELATION = "#%s"
)

const (
	ELLIPSIS = "..."
)

const (
	USER = "user"
)

// Object -
type Object struct {
	Namespace string
	ID        string
}

// UserSet -
type UserSet struct {
	Object   Object
	Relation Relation
}

// User -
type User struct {
	UserSet UserSet
	ID      string
}

// String -
func (u User) String() string {
	if u.IsUser() {
		return fmt.Sprintf("%s", u.ID)
	}
	return fmt.Sprintf("%s"+RELATION, fmt.Sprintf(OBJECT, u.UserSet.Object.Namespace, u.UserSet.Object.ID), u.UserSet.Relation)
}

// IsUser -
func (u User) IsUser() bool {
	if u.ID != "" {
		return true
	}
	return false
}

// Equals -
func (u User) Equals(v interface{}) bool {
	uv, ok := v.(User)
	if !ok {
		return false
	}
	if u.IsUser() {
		return u.ID == uv.ID
	}
	return uv.UserSet.Relation == u.UserSet.Relation && uv.UserSet.Object.ID == u.UserSet.Object.ID && uv.UserSet.Object.Namespace == u.UserSet.Object.Namespace
}

// Tuple -
type Tuple struct {
	Object   Object
	Relation string
	User     User
}

// String -
func (r Tuple) String() string {
	object := fmt.Sprintf(OBJECT, r.Object.Namespace, r.Object.ID)
	relation := fmt.Sprintf(RELATION, r.Relation)
	return object + relation + "@" + r.User.String()
}

// ConvertUser -
func ConvertUser(v string) User {
	parts := strings.Split(v, "#")
	if len(parts) != 2 {
		return User{
			ID: v,
		}
	}

	innerParts := strings.Split(parts[0], ":")
	if len(innerParts) != 2 {
		return User{
			ID: v,
		}
	}

	return User{
		UserSet: UserSet{
			Object: Object{
				Namespace: innerParts[0],
				ID:        innerParts[1],
			},
			Relation: Relation(parts[1]),
		},
	}
}

// ConvertObject -
func ConvertObject(v string) (Object, error) {
	obj := strings.Split(v, ":")
	if len(obj) < 1 {
		return Object{}, errors.New("input is not suitable for the object")
	}
	return Object{
		Namespace: obj[0],
		ID:        obj[1],
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
