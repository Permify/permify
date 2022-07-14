package utils

// IPagination -
type IPagination interface {
	Get() *Pagination
	GetPage() int
	GetLimit() int
}

// Pagination -
type Pagination struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

// Get -
func (p *Pagination) Get() *Pagination {
	return p
}

// GetPage -
func (p *Pagination) GetPage() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return p.Page
}

// GetLimit -
func (p *Pagination) GetLimit() int {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	return p.Limit
}
