// Package servicebus Websocket Server listening for connections to the pipeline
// - manage authentication
// - manage JWT encryption and signing when used
// - store and pass messages along the pipeline
package servicebus

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// StartServiceBus start listening for incoming connections and messages
// This returns after listening is established
// host contains the hostname and port, default is localhost:9678
// authmap is a map from pluginID to authorization token
func StartServiceBus(host string, authMap map[string]string) {
	if host == "" {
		host = "localhost:9678"
	}

	router := mux.NewRouter()
	router.HandleFunc("/channel/{ChannelID}/{Stage}", ServeChannel)
	router.HandleFunc("/echo", ServeEcho)
	router.HandleFunc("/", ServeHome)
	for pid, token := range authMap {
		AddAuthToken(pid, token)
	}

	go func() {
		err := http.ListenAndServe(host, router)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	// time.Sleep(time.Second)
}
