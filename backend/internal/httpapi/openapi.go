package httpapi

import (
	"io/fs"
	"net/http"
	"strconv"

	"brillio/backend"
)

// contentTypeYAML is the Content-Type of the OpenAPI document. RFC 9512
// registers application/yaml with no charset parameter.
const contentTypeYAML = "application/yaml"

// openAPIDocument is the embedded OpenAPI document (docs/SPEC.md section
// 5.7), read once at startup.
var openAPIDocument = func() []byte {
	b, err := fs.ReadFile(backend.FS, backend.OpenAPIPath)
	if err != nil {
		// The build fails when the file is missing, so this cannot happen.
		panic(err)
	}
	return b
}()

// openapi serves GET /api/openapi.yaml: the embedded document, byte for byte.
func (a *api) openapi(w http.ResponseWriter, _ *http.Request) {
	h := w.Header()
	h.Set("Content-Type", contentTypeYAML)
	h.Set("Content-Length", strconv.Itoa(len(openAPIDocument)))
	// A failed write means the client has gone away; there is nothing to do.
	_, _ = w.Write(openAPIDocument)
}
