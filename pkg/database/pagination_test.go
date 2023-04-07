package database

import (
	`testing`
)

func TestNewPagination(t *testing.T) {
	tests := []struct {
		name      string
		opts      []Option
		wantSize  uint32
		wantToken string
	}{
		{
			name:      "Default size",
			opts:      []Option{},
			wantSize:  _defaultPageSize,
			wantToken: "",
		},
		{
			name: "Custom size and token",
			opts: []Option{
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
	if p.PageSize() != _defaultPageSize {
		t.Errorf("Expected default page size of %d, but got %d", _defaultPageSize, p.PageSize())
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
