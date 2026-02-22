package frontend

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSPARoute(t *testing.T) {
	tests := []struct {
		path   string
		expect bool
	}{
		{"/", false},
		{".", false},
		{"/lists/abc", true},
		{"/lists/a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d", true},
		{"/search", true},
		{"/options", true},
		{"/api/foo", false},
		{"/metrics", false},
		{"/wallpapers/x", false},
		{"/index.html", false},
		{"/assets/main.js", false},
		{"/favicon.ico", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isSPARoute(tt.path)
			assert.Equal(t, tt.expect, got, "isSPARoute(%q)", tt.path)
		})
	}
}

func TestSPAFallback_ServesIndexHTMLForClientRoute(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.html")
	require.NoError(t, os.WriteFile(indexPath, []byte("<html>app</html>"), 0644))

	fs := http.FileServer(http.Dir(dir))
	handler := spaFallback(dir, fs)

	req := httptest.NewRequest(http.MethodGet, "http://test/lists/foo-id", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<html>app</html>")
}

func TestSPAFallback_ServesFileWhenExists(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "lists"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "lists", "some-file"), []byte("static"), 0644))

	fs := http.FileServer(http.Dir(dir))
	handler := spaFallback(dir, fs)

	req := httptest.NewRequest(http.MethodGet, "http://test/lists/some-file", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "static", rec.Body.String())
}
