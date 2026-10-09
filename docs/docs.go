package docs

import (
	_ "embed"
)

// OpenAPIJSON holds the raw OpenAPI 3.0 JSON specification for the Permify API.
//
//go:embed api-reference/openapi.json
var OpenAPIJSON []byte
