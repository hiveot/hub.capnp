package msgbus

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// ServeHome a home page if no path is given
func ServeHome(w http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if request.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	logrus.Infof("ServeHome")
	http.ServeFile(w, request, "../status/index.html")
}
