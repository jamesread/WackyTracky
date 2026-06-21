// Package apidocs serves machine-readable API discovery documents (OpenAPI, llms.txt).
package apidocs

import (
	"embed"
	"net/http"
)

//go:embed openapi.yaml llms.txt
var files embed.FS

func readOnlyGet(w http.ResponseWriter, r *http.Request, contentType, name string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := files.ReadFile(name)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		_, _ = w.Write(data)
	}
}

// OpenAPIHandler serves the embedded OpenAPI 3.1 spec at /openapi.
func OpenAPIHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readOnlyGet(w, r, "application/yaml; charset=utf-8", "openapi.yaml")
	})
}

// LLMsTxtHandler serves llms.txt at /llms.txt.
func LLMsTxtHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readOnlyGet(w, r, "text/plain; charset=utf-8", "llms.txt")
	})
}
