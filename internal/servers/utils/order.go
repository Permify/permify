package utils

import (
	"github.com/Permify/permify/pkg/helper"
)

// IOrder -
type IOrder interface {
	Get() *Order
	GetOrderBy() string
	GetSortBy() string
}

// Order -
type Order struct {
	OrderBy string `query:"order_by"`
	SortBy  string `query:"sort_by"`
}

// Get -
func (o *Order) Get() *Order {
	return o
}

// GetOrderBy -
func (o *Order) GetOrderBy() string {
	return o.OrderBy
}

// GetSortBy -
func (o *Order) GetSortBy() string {
	if !helper.InArray(o.SortBy, []string{"asc", "desc"}) {
		o.SortBy = "asc"
	}
	return o.SortBy
}
