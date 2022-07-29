package rules

import (
	"errors"
	"strings"
)

// IsObject -
func IsObject(value interface{}) error {
	s, _ := value.(string)
	obj := strings.Split(s, ":")
	if len(obj) < 2 {
		return errors.New("value is not suitable for the object")
	}
	return nil
}
