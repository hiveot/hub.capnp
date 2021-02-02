// Package servicebus Websocket Server listening for connections to the pipeline
// - manage authentication
// - manage JWT encryption and signing when used
// - store and pass messages along the pipeline
package servicebus

// DefaultHost with listening address and port
const DefaultHost = "localhost:9678"

// StartServiceBus start listening for incoming connections and messages.
// This returns after listening is established
// host contains the hostname and port
// clientAuth contains the client authorization tokens
func StartServiceBus(host string, clientAuth map[string]string) *ChannelServer {
	if host == "" {
		host = DefaultHost
	}

	srv := NewChannelServer()

	// ServeChannel handles incoming channel connections for pub or sub
	router := srv.Start(host)
	for pid, token := range clientAuth {
		srv.AddAuthToken(pid, token)
	}
	// ServeHome provides a status view
	router.HandleFunc("/", ServeHome)

	// time.Sleep(time.Second)
	return srv
}
