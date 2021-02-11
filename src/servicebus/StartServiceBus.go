// Package servicebus Websocket Server listening for connections to the service bus
package servicebus

import (
	"io/ioutil"
	"net"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/src/lib"
)

// DefaultHost with listening address and port
const DefaultHost = "localhost:9678"

// DefaultServerCert contains the server certificate file
const (
	CaCertFile     = "ca.crt"
	CaKeyFile      = "ca.key"
	ServerCertFile = "gateway.crt"
	ServerKeyFile  = "gateway.key"
	ClientCertFile = "client.crt"
	ClientKeyFile  = "client.key"
)

// const DefaultPubKey = "server.pub"

// StartServiceBus start listening for incoming connections and messages.
// This returns after listening is established
// - host contains the hostname and optionally port. Default is 9678 (WOST)
func StartServiceBus(host string) (*ChannelServer, error) {
	var router *mux.Router
	var err error

	if host == "" {
		host = DefaultHost
	}
	srv := NewChannelServer()

	router, err = srv.Start(host)

	// ServeHome provides a status view
	router.HandleFunc("/", ServeHome)

	// time.Sleep(time.Second)
	return srv, err
}

// StartTLSServiceBus start listening for incoming connections and messages over TLS.
// The certFolder contains the certificates for using TLS
// If no certificate is found in certFolder they will be generated.
// This returns after listening is established
// - host contains the hostname and optionally port. Default is 9678 (WOST)
// - certFolder is the folder for ca, server and client certificates
func StartTLSServiceBus(host string, certFolder string) (*ChannelServer, error) {
	var router *mux.Router
	var err error

	if host == "" {
		host = DefaultHost
	}
	srv := NewChannelServer()

	if certFolder == "" {
		router, err = srv.Start(host)
	} else {
		logrus.Warningf("StartServiceBus: Starting server using TLS")
		caCertPath := path.Join(certFolder, CaCertFile)
		caKeyPath := path.Join(certFolder, CaKeyFile)
		serverCertPath := path.Join(certFolder, ServerCertFile)
		serverKeyPath := path.Join(certFolder, ServerKeyFile)
		clientCertPath := path.Join(certFolder, ClientCertFile)
		clientKeyPath := path.Join(certFolder, ClientKeyFile)

		// err := httpscerts.Check("cert.pem", "key.pem")

		_, err = os.Stat(serverCertPath)
		if os.IsNotExist(err) {
			logrus.Warningf("StartServiceBus: Certificates not found. Generating new certificate files in %s", certFolder)
			caCertPEM, caKeyPEM := lib.CreateWoSTCA()
			hostname, port, err := net.SplitHostPort(host)
			_ = port
			if err != nil {
				return srv, err
			}
			// Certificate should not contain the port
			serverCertPEM, serverKeyPEM, _ := lib.CreateGatewayCert(caCertPEM, caKeyPEM, hostname)
			clientCertPEM, clientKeyPEM, _ := lib.CreateClientCert(caCertPEM, caKeyPEM, hostname)

			ioutil.WriteFile(caCertPath, caCertPEM, 0644)
			ioutil.WriteFile(caKeyPath, caKeyPEM, 0600)
			ioutil.WriteFile(serverKeyPath, serverKeyPEM, 0600)
			ioutil.WriteFile(serverCertPath, serverCertPEM, 0644)
			ioutil.WriteFile(clientKeyPath, clientKeyPEM, 0600)
			ioutil.WriteFile(clientCertPath, clientCertPEM, 0644)
		} else if err != nil {
			logrus.Errorf("StartServiceBus: Unable to open server certificate file %s: %s", serverCertPath, err)
			logrus.Fatal("Stopping")
		} else {
			logrus.Infof("StartServiceBus: Using server certificate file %s", serverCertPath)
		}
		router, err = srv.StartTLS(host, caCertPath, serverCertPath, serverKeyPath)
	}

	// ServeHome provides a status view
	router.HandleFunc("/", ServeHome)

	// time.Sleep(time.Second)
	return srv, err
}
