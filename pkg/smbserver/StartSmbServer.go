package smbserver

import (
	"io/ioutil"
	"net"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/certsetup"
	"github.com/wostzone/gateway/pkg/config"
)

// Start starts the built-in lightweigth message bus and listens for incoming connections and messages.
// This returns after listening is established
// - hostPort contains the hostname and port. Default is 9678 (WOST)
func Start(hostPort string) (*ServeSmbus, error) {
	var router *mux.Router
	var err error
	logrus.Warningf("Starting message bus server no TLS")

	if hostPort == "" {
		hostPort = config.DefaultSmbHost
	}
	srv := NewServeMsgBus()
	router, err = srv.Start(hostPort)

	// ServeHome provides a status view
	router.HandleFunc("/", ServeHome)

	// let server start up
	time.Sleep(time.Millisecond)
	return srv, err
}

// StartTLS start listening for incoming connections and messages over TLS.
// The certFolder contains the certificates for using TLS
// If no certificate is found in certFolder they will be generated.
// This returns after listening is established
// - host contains the hostname and optionally port. Default is 9678 (WOST)
// - certFolder is the folder for ca, server and client certificates. Default is ./certs
func StartTLS(host string, certFolder string) (*ServeSmbus, error) {
	var router *mux.Router
	var err error

	if host == "" {
		host = config.DefaultSmbHost
	}
	if certFolder == "" {
		certFolder = config.DefaultCertsFolder
	}
	srv := NewServeMsgBus()

	logrus.Warningf("Starting message bus server using TLS on %s", host)
	caCertPath := path.Join(certFolder, certsetup.CaCertFile)
	caKeyPath := path.Join(certFolder, certsetup.CaKeyFile)
	serverCertPath := path.Join(certFolder, certsetup.ServerCertFile)
	serverKeyPath := path.Join(certFolder, certsetup.ServerKeyFile)
	clientCertPath := path.Join(certFolder, certsetup.ClientCertFile)
	clientKeyPath := path.Join(certFolder, certsetup.ClientKeyFile)

	// err := httpscertsetup.Check("cert.pem", "key.pem")

	_, err = os.Stat(serverCertPath)
	if os.IsNotExist(err) {
		logrus.Warningf("Certificates not found. Generating new certificate files in %s", certFolder)
		caCertPEM, caKeyPEM := certsetup.CreateWoSTCA()
		hostname, port, err := net.SplitHostPort(host)
		_ = port
		if err != nil {
			return srv, err
		}
		// Certificate should not contain the port
		serverCertPEM, serverKeyPEM, _ := certsetup.CreateGatewayCert(caCertPEM, caKeyPEM, hostname)
		clientCertPEM, clientKeyPEM, _ := certsetup.CreateClientCert(caCertPEM, caKeyPEM, hostname)

		ioutil.WriteFile(caCertPath, caCertPEM, 0644)
		ioutil.WriteFile(caKeyPath, caKeyPEM, 0600)
		err = ioutil.WriteFile(serverKeyPath, serverKeyPEM, 0600)
		if err != nil {
			logrus.Errorf("Error creating certificates: %s", err)
			return nil, err
		}
		ioutil.WriteFile(serverCertPath, serverCertPEM, 0644)
		ioutil.WriteFile(clientKeyPath, clientKeyPEM, 0600)
		ioutil.WriteFile(clientCertPath, clientCertPEM, 0644)
	} else if err != nil {
		logrus.Errorf("Unable to open server certificate file %s: %s", serverCertPath, err)
		logrus.Fatal("Stopping")
	} else {
		logrus.Infof("Using server certificate file %s", serverCertPath)
	}
	router, err = srv.StartTLS(host, caCertPath, serverCertPath, serverKeyPath)

	// ServeHome provides a status view
	router.HandleFunc("/", ServeHome)

	// time.Sleep(time.Second)
	return srv, err
}

// StartSmbServer Main entry point to start the Simple Message Bus server with
// the given gateway configuration
func StartSmbServer(gwConfig *config.GatewayConfig) (*ServeSmbus, error) {
	var server *ServeSmbus
	var err error
	if gwConfig.Messenger.CertFolder != "" {
		server, err = StartTLS(gwConfig.Messenger.HostPort, gwConfig.Messenger.CertFolder)
	} else {
		server, err = Start(gwConfig.Messenger.HostPort)
	}
	return server, err
}
