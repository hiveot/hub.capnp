package servicebus

import "net/http"

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
	http.ServeFile(w, request, "home.html")
}
