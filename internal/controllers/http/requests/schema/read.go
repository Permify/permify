package schema

import (
	`github.com/Permify/permify/internal/utils`
)

// ReadRequest -
type ReadRequest struct {
	/**
	 * PathParams
	 */
	PathParams struct {
		SchemaVersion utils.Version `param:"schema_version"`
	}

	/**
	 * QueryParams
	 */
	QueryParams struct{}

	/**
	 * Body
	 */
	Body struct {
	}
}

// Validate -
func (r ReadRequest) Validate() (err error) {
	return nil
}
