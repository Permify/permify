package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEntityRef(t *testing.T) {
	t.Parallel()
	e, err := ParseEntityRef("document:1")
	require.NoError(t, err)
	assert.Equal(t, "document", e.GetType())
	assert.Equal(t, "1", e.GetId())
}

func TestParseEntityRef_Invalid(t *testing.T) {
	t.Parallel()
	_, err := ParseEntityRef("nocolon")
	require.Error(t, err)
}

func TestParseSubjectRef(t *testing.T) {
	t.Parallel()
	s, err := ParseSubjectRef("user:alice")
	require.NoError(t, err)
	assert.Equal(t, "user", s.GetType())
	assert.Equal(t, "alice", s.GetId())
}
