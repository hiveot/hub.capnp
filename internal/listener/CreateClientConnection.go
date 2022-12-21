package listener

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/svcconfig"
)

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
	var clientCertList []tls.Certificate
	var checkServerCert bool
	caCertPool := x509.NewCertPool()

	// Use CA certificate for server authentication if it exists
	if caCert == nil {
		// No CA certificate so no client authentication either
		logrus.Infof("destination '%s'. No CA certificate. InsecureSkipVerify used", address)
		checkServerCert = false
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

	// setup the tls client authentication
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
	}
	return conn, err
}

// CreateTLSClient2 wraps a net dial into TLS
//
// This listener accepts a client certificate for client authentication and a server CA certificate
// to verify the server connection.
//
//	 lis TCP listener
//	 clientCert is the client certificate to authenticate with. Use nil to not use client authentication
//		caCert is the CA certificate used to verify the server authenticity. Use nil if server auth is not yet established.
func CreateTLSClient2(conn net.Conn, clientCert *tls.Certificate, caCert *x509.Certificate) (*tls.Conn, error) {
	var clientCertList = make([]tls.Certificate, 0)
	var checkServerCert bool
	caCertPool := x509.NewCertPool()

	// Use CA certificate for server authentication if it exists
	if caCert == nil {
		// No CA certificate so no client authentication either
		checkServerCert = false
	} else {
		caCertPool.AddCert(caCert)
		checkServerCert = true

		opts := x509.VerifyOptions{
			Roots:     caCertPool,
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}
		x509Cert, err := x509.ParseCertificate(clientCert.Certificate[0])
		_, err = x509Cert.Verify(opts)
		if err != nil {
			logrus.Errorf("certificate verfication failed: %s", err)
			return nil, err
		}
	}

	// setup the tls client authentication
	clientCertList = append(clientCertList, *clientCert)

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       clientCertList,
		InsecureSkipVerify: !checkServerCert,
		ServerName:         "HiveOT Hub",
	}

	// finally, connect
	tlsConn := tls.Client(conn, tlsConfig)
	return tlsConn, nil
}
