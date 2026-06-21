package frontend

import (
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func init() {
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
}

func findWebUIDir() string {
	directoriesToSearch := []string{
		"../frontend/dist/",
		"/app/webui/",
		"../../wacky-tracky-client-html5/src/dist/",
		"../../../wacky-tracky-client-html5/src/dist/",
		"/usr/share/wacky-tracky/frontend/",
	}

	for i := 0; i < len(directoriesToSearch); i++ {
		if _, err := os.Stat(directoriesToSearch[i]); err == nil {
			return directoriesToSearch[i]
		}
	}

	log.Warnf("WebUI directory not found in any of the expected locations. Defaulting to ./webui")

	return "./webui"
}

var reservedPathPrefixes = []string{"/api", "/metrics", "/wallpapers", "/mcp"}

func hasReservedPrefix(p string) bool {
	for _, prefix := range reservedPathPrefixes {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

func isReservedPath(p string) bool {
	switch path.Clean(p) {
	case "/openapi", "/llms.txt":
		return true
	default:
		return hasReservedPrefix(p)
	}
}

// isSPARoute returns true for paths that are client-side routes (no static file).
func isSPARoute(p string) bool {
	p = path.Clean(p)
	if p == "." || p == "/" || path.Ext(p) != "" || isReservedPath(p) {
		return false
	}
	return true
}

func pathUnderWebUI(localPath, webuiDir string) bool {
	abs, _ := filepath.Abs(localPath)
	webuiAbs, _ := filepath.Abs(webuiDir)
	if abs == "" || webuiAbs == "" {
		return false
	}
	return abs == webuiAbs || strings.HasPrefix(abs, webuiAbs+string(filepath.Separator))
}

func regularFileExists(localPath string) bool {
	f, err := os.Stat(localPath)
	return err == nil && f != nil && !f.IsDir()
}

func getOrHead(r *http.Request) bool {
	return r.Method == http.MethodGet || r.Method == http.MethodHead
}

// isServiceWorkerAsset reports paths that must not be cached by intermediaries or
// an old browser service worker, so updates and auth cookies stay in sync.
func isServiceWorkerAsset(p string) bool {
	switch path.Clean(p) {
	case "/sw.js", "/pwa-sw.js", "/registerSW.js":
		return true
	default:
		return strings.HasPrefix(path.Clean(p), "/workbox-")
	}
}

func noStoreServiceWorkerAssets(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if getOrHead(r) && isServiceWorkerAsset(r.URL.Path) {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		}
		next.ServeHTTP(w, r)
	})
}

// spaFallback wraps a file server and serves index.html for GET/HEAD when the
// path looks like a client route and no file exists (e.g. /lists/..., /search).
func spaFallback(webuiDir string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !getOrHead(r) {
			next.ServeHTTP(w, r)
			return
		}
		if !isSPARoute(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		localPath := filepath.Join(webuiDir, filepath.FromSlash(strings.TrimPrefix(path.Clean(r.URL.Path), "/")))
		if !pathUnderWebUI(localPath, webuiDir) {
			next.ServeHTTP(w, r)
			return
		}
		if regularFileExists(localPath) {
			next.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(webuiDir, "index.html"))
	})
}

func GetNewHandler() http.Handler {
	webuiDir := findWebUIDir()

	log.WithFields(log.Fields{
		"webuiDir": webuiDir,
	}).Infof("WebUI directory found")

	fs := http.FileServer(http.Dir(webuiDir))
	return noStoreServiceWorkerAssets(spaFallback(webuiDir, fs))
}
