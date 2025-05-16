package frontend

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func findWebUIDir() string {
	directoriesToSearch := []string{
		"../webui/",
		"../../wacky-tracky-client-html5/src/dist/",
		"../../../wacky-tracky-client-html5/src/dist/",
		"/usr/share/wacky-tracky/webui/",
	}

	for i := 0; i < len(directoriesToSearch); i++ {
		if _, err := os.Stat(directoriesToSearch[i]); err == nil {
			return directoriesToSearch[i]
		}
	}

	return "./webui"
}

func GetNewHandler() http.Handler {
	webuiDir := findWebUIDir()

	log.WithFields(log.Fields{
		"webuiDir": webuiDir,
	}).Infof("WebUI directory found")

	return http.FileServer(http.Dir(webuiDir))
}
