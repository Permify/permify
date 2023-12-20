package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsts(t *testing.T) {
	tests := []struct {
		name   Engine
		result string
	}{
		{
			name:   "postgres",
			result: "postgres",
		},
		{
			name:   "memory",
			result: "memory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name.String(), func(t *testing.T) {
			assert.Equal(t, tt.result, tt.result)
		})
	}
}
