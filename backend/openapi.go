// Package backend embeds the files that live at the root of the backend
// module. It exists for openapi.yaml: go:embed cannot reach a parent
// directory, so the document is embedded here, next to go.mod where
// docs/SPEC.md section 3 puts it, and internal/httpapi serves it. It imports
// nothing internal, so any package may import it.
package backend

import "embed"

// OpenAPIPath is the path of the OpenAPI document within FS.
const OpenAPIPath = "openapi.yaml"

// FS holds the OpenAPI 3.1 document that GET /api/openapi.yaml serves
// (docs/SPEC.md section 5.7).
//
//go:embed openapi.yaml
var FS embed.FS
