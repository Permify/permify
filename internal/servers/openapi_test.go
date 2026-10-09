package servers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/require"
)

func TestRegisterOpenAPIHandlers(t *testing.T) {
	mux := runtime.NewServeMux()
	err := RegisterOpenAPIHandlers(mux)
	require.NoError(t, err)

	t.Run("GET /openapi.json returns 200 and valid OpenAPI JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var data map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &data)
		require.NoError(t, err)
		require.Equal(t, "3.0.0", data["openapi"])
		info, ok := data["info"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "Permify API", info["title"])
	})

	t.Run("GET /unregistered returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/unregistered", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
	})
}
