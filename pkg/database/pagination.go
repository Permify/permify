package database

// PaginationOption - Option type
type PaginationOption func(*Pagination)

// Size -
func Size(size uint32) PaginationOption {
	return func(c *Pagination) {
		c.size = size
	}
}

// Token -
func Token(token string) PaginationOption {
	return func(c *Pagination) {
		c.token = token
	}
}

// Pagination -
type Pagination struct {
	size  uint32
	token string
}

// NewPagination -
func NewPagination(opts ...PaginationOption) Pagination {
	pagination := &Pagination{}

	// Custom options
	for _, opt := range opts {
		opt(pagination)
	}

	return *pagination
}

// PageSize -
func (p Pagination) PageSize() uint32 {
	return p.size
}

// Token -
func (p Pagination) Token() string {
	return p.token
}

// CursorPaginationOption - Option type
type CursorPaginationOption func(*CursorPagination)

// CursorPagination -
type CursorPagination struct {
	cursor string
	sort   string
	limit  uint32
}

// NewCursorPagination -
func NewCursorPagination(opts ...CursorPaginationOption) CursorPagination {
	pagination := &CursorPagination{}

	// Custom options
	for _, opt := range opts {
		opt(pagination)
	}

	return *pagination
}

// Cursor -
func Cursor(cursor string) CursorPaginationOption {
	return func(c *CursorPagination) {
		c.cursor = cursor
	}
}

// Sort -
func Sort(sort string) CursorPaginationOption {
	return func(c *CursorPagination) {
		c.sort = sort
	}
}

// Limit -
func Limit(limit uint32) CursorPaginationOption {
	return func(c *CursorPagination) {
		c.limit = limit
	}
}

// Cursor -
func (p CursorPagination) Cursor() string {
	return p.cursor
}

// Sort -
func (p CursorPagination) Sort() string {
	return p.sort
}

// Limit -
func (p CursorPagination) Limit() uint32 {
	return p.limit
}
