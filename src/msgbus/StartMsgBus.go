package msgbus

import (
	"io/ioutil"
	"net"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/src/certs"
)

// const DefaultPubKey = "server.pub"

// Start starts the built-in message bus and listens for incoming connections and messages.
// This returns after listening is established
// - hostPort contains the hostname and port. Default is 9678 (WOST)
func Start(hostPort string) (*ServeMsgBus, error) {
	var router *mux.Router
	var err error
	logrus.Warningf("Start: Starting message bus server no TLS")

	if hostPort == "" {
		hostPort = DefaultMsgBusHost
	}
	srv := NewServeMsgBus()
	router, err = srv.Start(hostPort)

	// ServeHome provides a status view
	router.HandleFunc("/", ServeHome)

	// time.Sleep(time.Second)
	return srv, err
}

// StartTLS start listening for incoming connections and messages over TLS.
// The certFolder contains the certificates for using TLS
// If no certificate is found in certFolder they will be generated.
// This returns after listening is established
// - host contains the hostname and optionally port. Default is 9678 (WOST)
// - certFolder is the folder for ca, server and client certificates
func StartTLS(host string, certFolder string) (*ServeMsgBus, error) {
	var router *mux.Router
	var err error

	if host == "" {
		host = DefaultMsgBusHost
	}
	srv := NewServeMsgBus()

	logrus.Warningf("StartTLS: Starting message bus server using TLS")
	caCertPath := path.Join(certFolder, CaCertFile)
	caKeyPath := path.Join(certFolder, CaKeyFile)
	serverCertPath := path.Join(certFolder, ServerCertFile)
	serverKeyPath := path.Join(certFolder, ServerKeyFile)
	clientCertPath := path.Join(certFolder, ClientCertFile)
	clientKeyPath := path.Join(certFolder, ClientKeyFile)

	// err := httpscerts.Check("cert.pem", "key.pem")

	_, err = os.Stat(serverCertPath)
	if os.IsNotExist(err) {
		logrus.Warningf("StartTLS: Certificates not found. Generating new certificate files in %s", certFolder)
		caCertPEM, caKeyPEM := certs.CreateWoSTCA()
		hostname, port, err := net.SplitHostPort(host)
		_ = port
		if err != nil {
			return srv, err
		}
		// Certificate should not contain the port
		serverCertPEM, serverKeyPEM, _ := certs.CreateGatewayCert(caCertPEM, caKeyPEM, hostname)
		clientCertPEM, clientKeyPEM, _ := certs.CreateClientCert(caCertPEM, caKeyPEM, hostname)

		ioutil.WriteFile(caCertPath, caCertPEM, 0644)
		ioutil.WriteFile(caKeyPath, caKeyPEM, 0600)
		ioutil.WriteFile(serverKeyPath, serverKeyPEM, 0600)
		ioutil.WriteFile(serverCertPath, serverCertPEM, 0644)
		ioutil.WriteFile(clientKeyPath, clientKeyPEM, 0600)
		ioutil.WriteFile(clientCertPath, clientCertPEM, 0644)
	} else if err != nil {
		logrus.Errorf("StartTLS: Unable to open server certificate file %s: %s", serverCertPath, err)
		logrus.Fatal("Stopping")
	} else {
		logrus.Infof("StartTLS: Using server certificate file %s", serverCertPath)
	}
	router, err = srv.StartTLS(host, caCertPath, serverCertPath, serverKeyPath)

	// ServeHome provides a status view
	router.HandleFunc("/", ServeHome)

	// time.Sleep(time.Second)
	return srv, err
}
