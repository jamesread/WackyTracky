package frontend

import (
	"mime"
	"net/http"
	"os"

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

func GetNewHandler() http.Handler {
	webuiDir := findWebUIDir()

	log.WithFields(log.Fields{
		"webuiDir": webuiDir,
	}).Infof("WebUI directory found")

	return http.FileServer(http.Dir(webuiDir))
}
