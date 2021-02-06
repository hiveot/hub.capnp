// Package servicebus Websocket Server listening for connections to the pipeline
// - manage authentication
// - manage JWT encryption and signing when used
// - store and pass messages along the pipeline
package servicebus

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DefaultHost with listening address and port
const DefaultHost = "localhost:9678"

// DefaultServerCert contains the server certificate file
const (
	caCertFile     = "ca.crt"
	caKeyFile      = "ca.key"
	serverCertFile = "gateway.crt"
	serverKeyFile  = "gateway.key"
)

// const DefaultPubKey = "server.pub"

// StartServiceBus start listening for incoming connections and messages.
// If a certFolder is provided the server will listen using TLS instead HTTP
// If no certificate is found in certFolder, one will be generated.
//    caServer, server.crt, server.key
// This returns after listening is established
// - host contains the hostname and port
// - certFolder is the folder for server and client certificates when using TLS
// - clientAuth contains the client authorization tokens
func StartServiceBus(host string, certFolder string, clientAuth map[string]string) *ChannelServer {
	var router *mux.Router

	if host == "" {
		host = DefaultHost
	}
	srv := NewChannelServer()

	if certFolder == "" {
		router = srv.Start(host)
	} else {
		caCertPath := path.Join(certFolder, caCertFile)
		caKeyPath := path.Join(certFolder, caKeyFile)
		serverCertPath := path.Join(certFolder, serverCertFile)
		serverKeyPath := path.Join(certFolder, serverKeyFile)

		// err := httpscerts.Check("cert.pem", "key.pem")

		_, err := os.Stat(serverCertPath)
		if os.IsNotExist(err) {
			caCertPEM, caKeyPEM := CreateWoSTCA()
			serverCertPEM, serverKeyPEM, _ := CreateGatewayCert(caCertPEM, caKeyPEM, host)

			ioutil.WriteFile(caCertPath, caCertPEM, os.FileMode(os.O_CREATE))
			ioutil.WriteFile(caKeyPath, caKeyPEM, os.FileMode(os.O_CREATE))
			ioutil.WriteFile(serverKeyPath, serverKeyPEM, os.FileMode(os.O_CREATE))
			ioutil.WriteFile(serverCertPath, serverCertPEM, os.FileMode(os.O_CREATE))
		} else if err != nil {
			logrus.Errorf("Unable to open server certificate file %s: %s", serverCertPath, err)
			logrus.Fatal("Stopping")
		}
		router = srv.StartTLS(host, serverCertPath, serverKeyPath)
	}

	// ServeChannel handles incoming channel connections for pub or sub
	// router, clientTLSConf := srv.Start(host, serverCertPath, serverKeyPath)
	for pid, token := range clientAuth {
		srv.AddAuthToken(pid, token)
	}
	// ServeHome provides a status view
	router.HandleFunc("/", ServeHome)

	// time.Sleep(time.Second)
	return srv
}
