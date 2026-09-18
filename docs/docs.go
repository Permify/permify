package docs

import (
	_ "embed"
)

// OpenAPIJSON holds the OpenAPI 3.0 JSON specification.
//
//go:embed api-reference/openapi.json
var OpenAPIJSON []byte

// SwaggerJSON holds the Swagger 2.0 / OpenAPI 2.0 JSON specification.
//
//go:embed api-reference/apidocs.swagger.json
var SwaggerJSON []byte

// OpenAPIV2JSON holds the OpenAPI 2.0 JSON specification.
//
//go:embed api-reference/openapiv2/apidocs.swagger.json
var OpenAPIV2JSON []byte
