package servers

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/Permify/permify/docs"
)

// RegisterOpenAPIHandlers registers endpoints on the gRPC-Gateway ServeMux to expose
// OpenAPI and Swagger specifications over HTTP.
func RegisterOpenAPIHandlers(mux *runtime.ServeMux) error {
	endpoints := map[string][]byte{
		"/openapi.json":        docs.OpenAPIJSON,
		"/swagger.json":        docs.SwaggerJSON,
		"/docs/openapi.json":   docs.OpenAPIJSON,
		"/docs/swagger.json":   docs.SwaggerJSON,
		"/docs/openapiv2.json": docs.OpenAPIV2JSON,
	}

	for path, spec := range endpoints {
		content := spec
		err := mux.HandlePath(http.MethodGet, path, func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(content)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
