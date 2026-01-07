// Package api provides embedded OpenAPI specifications and Swagger UI.
package api

import (
	_ "embed"
)

//go:embed v1/openapi.yaml
var OpenAPIV1YAML []byte

//go:embed v1/openapi.json
var OpenAPIV1JSON []byte

//go:embed swagger-ui.html
var SwaggerUIHTML []byte
