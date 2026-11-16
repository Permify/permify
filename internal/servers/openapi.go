package servers

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/Permify/permify/docs"
)

// RegisterOpenAPIHandlers registers the OpenAPI endpoint on the gRPC-Gateway ServeMux
// to expose the API specification over HTTP.
func RegisterOpenAPIHandlers(mux *runtime.ServeMux) error {
	return mux.HandlePath(http.MethodGet, "/openapi.json", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(docs.OpenAPIJSON)
	})
}
