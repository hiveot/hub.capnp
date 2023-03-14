package hubclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
)

// ConnectToService creates a connection to a service using UDS
// using the convention that connection address = socketPath/service.socket
func ConnectToService(serviceName, socketFolder string) (net.Conn, error) {
	timeout := time.Second * 3
	socketPath := fmt.Sprintf("%s/%s.socket", socketFolder, serviceName)
	return net.DialTimeout("unix", socketPath, timeout)
}

// CreateClientConnection returns a client connected to the given address using
// tcp or websockets.
//
// This accepts a client certificate for client authentication and a server CA certificate
// to verify the server connection. If neither client nor CA certificate are provided TLS is not used.
// If the url schema is 'unix' then a local UDS is used and certificates are ignored.
//
// Note that when connecting to websockets, capnp needs a special media encoder to talk to the
// server.
//
//	fullUrl supports both tcp and wss for websocket connections. For example:
//	      unix://path/
//		  tcp://server:port/
//		  wss://server:port/ws
//
//	clientCert is the client certificate to authenticate with. Use nil to not use client authentication
//	caCert is the CA certificate used to verify the server authenticity. Use nil if server auth is not yet established.
func CreateClientConnection(
	fullUrl string, clientCert *tls.Certificate, caCert *x509.Certificate) (conn net.Conn, err error) {

	// var tlsConn *tls.Conn
	const timeout = time.Second * 3
	var tlsConfig *tls.Config
	var clientCertList []tls.Certificate = nil
	var checkServerCert bool
	caCertPool := x509.NewCertPool()

	// Use CA certificate for server authentication if it exists
	if caCert == nil {
		// No CA certificate so no client authentication either
		// logrus.Warningf("destination '%s'. No CA certificate. InsecureSkipVerify used", fullUrl)
		checkServerCert = false
	} else if clientCert == nil {
		// No CA certificate so no client authentication either
		// logrus.Warningf("No client certificate for connecting to '%s'. Client auth unavailable", fullUrl)
	} else {
		// CA certificate is provided
		logrus.Infof("Using TLS client auth. destination '%s'. CA certificate '%s'", fullUrl, caCert.Subject.CommonName)
		caCertPool.AddCert(caCert)
		checkServerCert = true

		opts := x509.VerifyOptions{
			Roots:     caCertPool,
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}
		if clientCert != nil {
			x509Cert, _ := x509.ParseCertificate(clientCert.Certificate[0])
			_, err = x509Cert.Verify(opts)
			if err != nil {
				logrus.Errorf("ConnectWithClientCert: certificate verfication failed: %s", err)
				return nil, err
			}
		}

		// setup for mutual tls client authentication
		if clientCert != nil {
			clientCertList = []tls.Certificate{*clientCert}
		}

		tlsConfig = &tls.Config{
			RootCAs:            caCertPool,
			Certificates:       clientCertList,
			InsecureSkipVerify: !checkServerCert,
		}
	}
	// default url to tcp://
	// tcp://adddr:port/path ->
	u, err := url.Parse(fullUrl)
	if err != nil {
		u = &url.URL{Scheme: "tcp", Host: fullUrl}
		fullUrl = "tcp://" + fullUrl
	}
	if u.Scheme == "ws" {
		// websocket no TLS
		dialer := ws.Dialer{Timeout: time.Second * 3}
		// TODO: check if a buffer is returned and return it to the pool if not nil
		conn, _, _, err = dialer.Dial(context.Background(), fullUrl)
	} else if u.Scheme == "wss" {
		// websocket with TLS - falls back to no TLS if tlsConfig is nil
		// TODO: check if a buffer is returned and return it to the pool if not nil
		dialer := ws.Dialer{TLSConfig: tlsConfig, Timeout: time.Second * 3}
		conn, _, _, err = dialer.Dial(context.Background(), fullUrl)
	} else if u.Scheme == "unix" {
		// Unix domain socket
		conn, err = net.DialTimeout(u.Scheme, u.Path, timeout)
	} else {
		// TCP socket
		if tlsConfig != nil {
			conn, err = tls.Dial(u.Scheme, u.Host, tlsConfig)
		} else {
			conn, err = net.DialTimeout(u.Scheme, u.Host, timeout)
		}
	}

	if err != nil {
		err = fmt.Errorf("unable to connect to '%s'. Error: %s", fullUrl, err)
		logrus.Error(err)
	} else {
		//logrus.Infof("connected to '%s'", fullUrl)
	}
	return conn, err
}
