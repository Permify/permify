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

	testCases := []struct {
		name         string
		path         string
		expectedCode int
		validateJSON func(t *testing.T, body []byte)
	}{
		{
			name:         "GET /openapi.json",
			path:         "/openapi.json",
			expectedCode: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var data map[string]any
				err := json.Unmarshal(body, &data)
				require.NoError(t, err)
				require.Equal(t, "3.0.0", data["openapi"])
				info, ok := data["info"].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "Permify API", info["title"])
			},
		},
		{
			name:         "GET /swagger.json",
			path:         "/swagger.json",
			expectedCode: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var data map[string]any
				err := json.Unmarshal(body, &data)
				require.NoError(t, err)
				require.Equal(t, "2.0", data["swagger"])
				info, ok := data["info"].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "Permify API", info["title"])
			},
		},
		{
			name:         "GET /docs/openapi.json",
			path:         "/docs/openapi.json",
			expectedCode: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var data map[string]any
				err := json.Unmarshal(body, &data)
				require.NoError(t, err)
				require.Equal(t, "3.0.0", data["openapi"])
			},
		},
		{
			name:         "GET /docs/swagger.json",
			path:         "/docs/swagger.json",
			expectedCode: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var data map[string]any
				err := json.Unmarshal(body, &data)
				require.NoError(t, err)
				require.Equal(t, "2.0", data["swagger"])
			},
		},
		{
			name:         "GET /docs/openapiv2.json",
			path:         "/docs/openapiv2.json",
			expectedCode: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var data map[string]any
				err := json.Unmarshal(body, &data)
				require.NoError(t, err)
				require.Equal(t, "2.0", data["swagger"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			require.Equal(t, tc.expectedCode, w.Code)
			require.Equal(t, "application/json", w.Header().Get("Content-Type"))
			if tc.validateJSON != nil {
				tc.validateJSON(t, w.Body.Bytes())
			}
		})
	}
}
