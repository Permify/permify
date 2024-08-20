package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPagination(t *testing.T) {
	tests := []struct {
		name      string
		opts      []PaginationOption
		wantSize  uint32
		wantToken string
	}{
		{
			name:      "Default size",
			opts:      []PaginationOption{},
			wantSize:  0,
			wantToken: "",
		},
		{
			name: "Custom size and token",
			opts: []PaginationOption{
				Size(50),
				Token("abc123"),
			},
			wantSize:  50,
			wantToken: "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPagination(tt.opts...)

			if p.size != tt.wantSize {
				t.Errorf("NewPagination() size = %d, want %d", p.size, tt.wantSize)
			}

			if p.token != tt.wantToken {
				t.Errorf("NewPagination() token = %s, want %s", p.token, tt.wantToken)
			}
		})
	}
}

func TestPagination(t *testing.T) {
	// Test default page size
	p := NewPagination()
	if p.PageSize() != 0 {
		t.Errorf("Expected default page size of %d, but got %d", 0, p.PageSize())
	}

	// Test custom page size
	p = NewPagination(Size(25))
	if p.PageSize() != 25 {
		t.Errorf("Expected page size of 25, but got %d", p.PageSize())
	}

	// Test token
	p = NewPagination(Token("my-token"))
	if p.Token() != "my-token" {
		t.Errorf("Expected token of 'my-token', but got '%s'", p.Token())
	}
}

func TestNoopContinuousToken(t *testing.T) {
	token := NewNoopContinuousToken()

	// Test Encode
	encodedToken := token.Encode()
	assert.Empty(t, encodedToken.String())

	// Test Decode
	decodedToken, err := encodedToken.Decode()
	assert.NoError(t, err)
	assert.Empty(t, decodedToken.(NoopContinuousToken).Value)

	// Test Encode and Decode
	assert.Empty(t, decodedToken.(NoopContinuousToken).Value)
}

func TestNewCursorPagination(t *testing.T) {
	tests := []struct {
		name       string
		opts       []CursorPaginationOption
		wantSort   string
		wantCursor string
	}{
		{
			name:       "Default size",
			opts:       []CursorPaginationOption{},
			wantSort:   "",
			wantCursor: "",
		},
		{
			name: "Custom size and token",
			opts: []CursorPaginationOption{
				Sort("entity_id"),
				Cursor("abc123"),
			},
			wantSort:   "entity_id",
			wantCursor: "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewCursorPagination(tt.opts...)

			if p.sort != tt.wantSort {
				t.Errorf("NewCursorPagination() size = %s, want %s", p.sort, tt.wantSort)
			}

			if p.cursor != tt.wantCursor {
				t.Errorf("NewCursorPagination() cursor = %s, want %s", p.cursor, tt.wantCursor)
			}
		})
	}
}

func TestCursorPagination(t *testing.T) {
	// Test default page size
	p := NewCursorPagination()
	if p.Sort() != "" {
		t.Errorf("Expected default sort empty string, but got %s", p.Sort())
	}

	// Test custom page size
	p = NewCursorPagination(Sort("entity_id"))
	if p.Sort() != "entity_id" {
		t.Errorf("Expected sort entity_id, but got %s", p.Sort())
	}

	// Test token
	p = NewCursorPagination(Cursor("my-cursor"))
	if p.Cursor() != "my-cursor" {
		t.Errorf("Expected token of 'my-cursor', but got '%s'", p.Cursor())
	}
}
