package smbserver

import (
	"io/ioutil"
	"net"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/certs"
	"github.com/wostzone/gateway/pkg/lib"
	"github.com/wostzone/gateway/pkg/messaging/smbus"
)

// const DefaultPubKey = "server.pub"

// SmbusConfig based on gateway configuration
type SmbusConfig struct {
	lib.GatewayConfig
}

// Start starts the built-in lightweigth message bus and listens for incoming connections and messages.
// This returns after listening is established
// - hostPort contains the hostname and port. Default is 9678 (WOST)
func Start(hostPort string) (*ServeSmbus, error) {
	var router *mux.Router
	var err error
	logrus.Warningf("Start: Starting message bus server no TLS")

	if hostPort == "" {
		hostPort = smbus.DefaultSmbusHost
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
// - certFolder is the folder for ca, server and client certificates
func StartTLS(host string, certFolder string) (*ServeSmbus, error) {
	var router *mux.Router
	var err error

	if host == "" {
		host = smbus.DefaultSmbusHost
	}
	srv := NewServeMsgBus()

	logrus.Warningf("StartTLS: Starting message bus server using TLS")
	caCertPath := path.Join(certFolder, certs.CaCertFile)
	caKeyPath := path.Join(certFolder, certs.CaKeyFile)
	serverCertPath := path.Join(certFolder, certs.ServerCertFile)
	serverKeyPath := path.Join(certFolder, certs.ServerKeyFile)
	clientCertPath := path.Join(certFolder, certs.ClientCertFile)
	clientKeyPath := path.Join(certFolder, certs.ClientKeyFile)

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
		err = ioutil.WriteFile(serverKeyPath, serverKeyPEM, 0600)
		if err != nil {
			logrus.Errorf("StartTLS: Error creating certificates: %s", err)
			return nil, err
		}
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

// StartSmbusServer Main entry point to start the Simple Message Bus server
func StartSmbusServer(config *SmbusConfig) (*ServeSmbus, error) {
	var server *ServeSmbus
	var err error
	if config.Messenger.UseTLS {
		server, err = StartTLS(config.Messenger.HostPort, config.Messenger.CertsFolder)
	} else {
		server, err = Start(config.Messenger.HostPort)
	}
	return server, err
}
