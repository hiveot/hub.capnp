package hubclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/gateway/config"
)

// DiscoveryGateway provides the gateway address
// TODO: add auto-discovery of the gateway
func DiscoveryGateway() string {
	address := config.DefaultGatewayAddress
	return address
}

// ConnectToHub provides a 'reconnectable' connection to the HiveOT Hub. It is intended to improve
// resiliency in cases where the connection with hub services can be broken, whether it is due to a
// network interruption or due to a service restart.
//
//	network is optional "unix" for UDS or "tcp" for TCP connections, default "" for auto discovery
//	address is optional path to UDS socket or address:port for tcp. "" for auto discover
//	clientCert is optional and only for use with gateway over TCP connections. nil for unauthenticated connections
//	caCert is optional, but recommended, to verify the gateway certificate and protect against MiM attack
func ConnectToHub(
	network, address string, clientCert *tls.Certificate, caCert *x509.Certificate) (conn net.Conn, err error) {

	// check if the resolver can be used
	if address == "" {
		_, err = os.Stat(hubapi.DefaultResolverAddress)
		if err == nil {
			address = hubapi.DefaultResolverAddress
			network = "unix"
		}
	}
	// determine the address to connect to
	if address == "" {
		if network == "unix" {
			address = hubapi.DefaultResolverAddress
		} else {
			network = "tcp"
			address = DiscoveryGateway()
		}
	}
	if network == "unix" {
		conn, err = net.DialTimeout(network, address, time.Second)
	} else {
		conn, err = CreateTLSClientConnection(network, address, clientCert, caCert)
	}
	if err != nil {
		logrus.Infof("connection failed. network='%s', address='%s'. Err=%s", network, address, err)
	} else {
		logrus.Infof("successfully connected. network='%s', address='%s'", network, address)
	}
	return conn, err
}

// CreateLocalClientConnection returns a local client connection for the given service.
//
// The service itself must listen on the unix domain socket for the service following the
// convention: {runFolder}/{serviceName}.socket
//
//	serviceName is the name of the service to connect to
//	runFolder is the folder containing sockets, or "" for default {home}/run
func CreateLocalClientConnection(serviceName string, runFolder string) (net.Conn, error) {
	if runFolder == "" {
		f := svcconfig.GetFolders("", false)
		runFolder = f.Run
	}
	svcAddress := filepath.Join(runFolder, serviceName+".socket")
	conn, err := net.DialTimeout("unix", svcAddress, time.Second)
	if err != nil {
		err = fmt.Errorf("Unable to connect to service socket '%s'. Is the service running?\n  Error: %s", svcAddress, err)
	}
	return conn, err
}

// CreateTLSClientConnection returns a TLS client connected to the given address
//
// This listener accepts a client certificate for client authentication and a server CA certificate
// to verify the server connection.
//
//	network is "unix" for unix domain sockets or tcp for TCP
//	address of the server to connect to in the form: "address:port"
//	clientCert is the client certificate to authenticate with. Use nil to not use client authentication
//	caCert is the CA certificate used to verify the server authenticity. Use nil if server auth is not yet established.
func CreateTLSClientConnection(network, address string, clientCert *tls.Certificate, caCert *x509.Certificate) (*tls.Conn, error) {
	var clientCertList []tls.Certificate = nil
	var checkServerCert bool
	caCertPool := x509.NewCertPool()

	// Use CA certificate for server authentication if it exists
	if caCert == nil {
		// No CA certificate so no client authentication either
		logrus.Warningf("destination '%s'. No CA certificate. InsecureSkipVerify used", address)
		checkServerCert = false
	} else if clientCert == nil {
		// No CA certificate so no client authentication either
		logrus.Warningf("No client certificate for connecting to '%s'. Client auth unavailable", address)
	} else {
		// CA certificate is provided
		logrus.Infof("destination '%s'. CA certificate '%s'",
			address, caCert.Subject.CommonName)
		caCertPool.AddCert(caCert)
		checkServerCert = true

		opts := x509.VerifyOptions{
			Roots:     caCertPool,
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}
		if clientCert != nil {
			x509Cert, err := x509.ParseCertificate(clientCert.Certificate[0])
			_, err = x509Cert.Verify(opts)
			if err != nil {
				logrus.Errorf("ConnectWithClientCert: certificate verfication failed: %s", err)
				return nil, err
			}
		}
	}

	// setup for mutual tls client authentication
	if clientCert != nil {
		clientCertList = []tls.Certificate{*clientCert}
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       clientCertList,
		InsecureSkipVerify: !checkServerCert,
	}

	// finally, connect
	conn, err := tls.Dial(network, address, tlsConfig)
	if err != nil {
		err = fmt.Errorf("Unable to connect to '%s'. Is the service running?\n  Error: %s", address, err)
		logrus.Error(err)
	} else {
		logrus.Infof("connected")
	}
	return conn, err
}
