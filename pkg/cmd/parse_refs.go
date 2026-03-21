package cmd

import (
	"fmt"
	"strings"

	basev1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// ParseEntityRef parses "type:id" into a base Entity.
func ParseEntityRef(s string) (*basev1.Entity, error) {
	s = strings.TrimSpace(s)
	typ, id, err := splitTypeID(s)
	if err != nil {
		return nil, fmt.Errorf("entity %q: %w", s, err)
	}
	return &basev1.Entity{Type: typ, Id: id}, nil
}

// ParseSubjectRef parses "type:id" into a base Subject (optional relation is not used by this helper).
func ParseSubjectRef(s string) (*basev1.Subject, error) {
	s = strings.TrimSpace(s)
	typ, id, err := splitTypeID(s)
	if err != nil {
		return nil, fmt.Errorf("subject %q: %w", s, err)
	}
	return &basev1.Subject{Type: typ, Id: id}, nil
}

func splitTypeID(s string) (typ, id string, err error) {
	if s == "" {
		return "", "", fmt.Errorf("expected non-empty type:id")
	}
	i := strings.Index(s, ":")
	if i <= 0 || i == len(s)-1 {
		return "", "", fmt.Errorf("expected type:id (e.g. user:1)")
	}
	return s[:i], s[i+1:], nil
}
