package apidocs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi", nil)
	OpenAPIHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/yaml; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.True(t, strings.HasPrefix(rec.Body.String(), "openapi: 3.1.0\n"))
}

func TestLLMsTxtHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/llms.txt", nil)
	LLMsTxtHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "/openapi")
	assert.Contains(t, rec.Body.String(), "/mcp")
	assert.Contains(t, rec.Body.String(), "list_lists")
}

func TestHandlersRejectPost(t *testing.T) {
	for _, h := range []http.Handler{OpenAPIHandler(), LLMsTxtHandler()} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	}
}
