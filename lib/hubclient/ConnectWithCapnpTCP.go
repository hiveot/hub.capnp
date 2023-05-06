package hubclient

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"net/url"
	"time"
)

// ConnectTCP returns a TCP network connection optionally using TLS
//
// This accepts a client certificate for client authentication and a server CA certificate
// to verify the server connection. If neither client nor CA certificate are provided TLS is not used.
// If the url schema is 'unix' then a local UDS is used and certificates are ignored.
//
//	fullUrl supports both tcp and uds connections. For example:
//	      unix://path/to/socket
//		  tcp://server:port/
//
//	clientCert is the client certificate to authenticate with. Use nil to not use client authentication
//	caCert is the CA certificate used to verify the server authenticity. Use nil if server auth is not yet established.
//
// This returns the network or tls connection or an error
func ConnectTCP(fullUrl string, clientCert *tls.Certificate, caCert *x509.Certificate) (conn net.Conn, err error) {

	const timeout = time.Second * 3
	var tlsConfig *tls.Config
	var clientCertList []tls.Certificate = nil
	var checkServerCert bool
	caCertPool := x509.NewCertPool()

	// use gateway discovery
	if fullUrl == "" {
		return nil, errors.New("missing URL")
		// this can add 0.6 MB of code so lets leave it up to the user whether this is needed
		//fullUrl = LocateHub("", searchTimeSec)
	}

	// Use CA certificate for server authentication if it exists
	if caCert == nil {
		// No CA certificate so no client authentication either
		// logrus.Warningf("destination '%s'. No CA certificate. InsecureSkipVerify used", fullUrl)
		checkServerCert = false
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

		// TLS setup for mutual tls client authentication
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
	// tcp://addr:port/path ->
	u, err := url.Parse(fullUrl)
	if err != nil {
		u = &url.URL{Scheme: "tcp", Host: fullUrl}
		fullUrl = "tcp://" + fullUrl
	}
	if u.Scheme == "unix" {
		// Support Unix domain socket. TLS is not needed.
		conn, err = net.DialTimeout(u.Scheme, u.Path, timeout)
	} else {
		// TCP socket
		if tlsConfig != nil {
			conn, err = tls.Dial(u.Scheme, u.Host, tlsConfig)
		} else {
			conn, err = net.DialTimeout(u.Scheme, u.Host, timeout)
		}
	}
	return conn, err
}

// ConnectWithCapnpTCP returns a capnp client connected to the capnp service with the given address using tcp.
//
// This accepts a client certificate for client authentication and a server CA certificate
// to verify the server connection. If neither client nor CA certificate are provided TLS is not used.
// If the url schema is 'unix' then a local UDS is used and certificates are ignored.
//
//	fullUrl supports both tcp and uds connections. For example:
//	      unix://path/to/socket
//		  tcp://server:port/
//
//	clientCert is the client certificate to authenticate with. Use nil to not use client authentication
//	caCert is the CA certificate used to verify the server authenticity. Use nil if server auth is not yet established.
//
// This returns the capnp client. To close the connection, Release the client.
func ConnectWithCapnpTCP(
	fullUrl string, clientCert *tls.Certificate, caCert *x509.Certificate) (
	client capnp.Client, err error) {

	var conn net.Conn
	conn, err = ConnectTCP(fullUrl, clientCert, caCert)

	if err != nil {
		err = fmt.Errorf("unable to connect to '%s'. Error: %s", fullUrl, err)
		logrus.Error(err)
	} else {
		// add a capnp transport for this connection and return a bootstrap client
		// the bootstrap client is a generic capnp client that can be used to send
		// capnp encoded messages.
		tp := rpc.NewStreamTransport(conn)
		rpcConn := rpc.NewConn(tp, nil)
		client = rpcConn.Bootstrap(context.Background())
	}
	return client, err
}
