package hubclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
	"net"
	"net/url"
	"strings"
	"time"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"capnproto.org/go/capnp/v3/rpc/transport"
	websocketcapnp "zenhack.net/go/websocket-capnp"
)

// ConnectToService using TLS and discovery.
//
// If fullURL is not provided this will attempt to auto-discover the Hub.
//
//  1. Unix socket at the default resolver path
//
//  2. Gateway TCP address at localhost:8883
//
//  3. DNS-SD for the hub gateway service (TODO)
//
//     fullURL is optional connection endpoint. Use "" to auto discover the resolver or gateway.
//func ConnectToService(fullUrl string, clientCert *tls.Certificate, caCert *x509.Certificate) (net.Conn, error) {
//	// use gateway discovery
//	if fullUrl == "" {
//		// give it 3 seconds
//		fullUrl = LocateHub("")
//	}
//
//	return ConnectToService(fullUrl, clientCert, caCert)
//}

// ConnectToHubClient returns the connection and capnp client of the gateway or resolver service.
//
// This client is special in that a request for a capability is dynamically forwarded to the actual service.
// If the client does not have the proper authentication type then the capability is not available and this fails
// with 'unimplemented'.
//
// Note that when connecting without client certificate to the gateway, Login must be called to authenticate.
//
// If fullURL is not provided this will attempt to auto-discover the Hub using 'LocateHub'
//  1. Unix socket at the default resolver path
//  2. DNS-SD for the hub gateway service
//
// This auto-discovers the gateway or default to 127.0.0.1:8883
func ConnectToHubClient(
	fullUrl string, searchTimeSec int, clientCert *tls.Certificate, caCert *x509.Certificate) (
	rpcCon *rpc.Conn, cap capnp.Client, err error) {

	var tp transport.Transport

	conn, err := ConnectToService(fullUrl, searchTimeSec, clientCert, caCert)
	if err != nil {
		return nil, cap, err
	}

	ctx := context.Background()
	// websockets use a capnp protocol en/decoder for its transport
	if strings.HasPrefix(fullUrl, "ws://") {
		// websocket without TLS for testing
		codec := websocketcapnp.NewCodec(conn, false)
		tp = transport.New(codec)
	} else if strings.HasPrefix(fullUrl, "wss://") {
		codec := websocketcapnp.NewCodec(conn, false)
		tp = transport.New(codec)
	} else {
		tp = rpc.NewStreamTransport(conn)
	}
	rpcConn := rpc.NewConn(tp, nil)
	hubClient := rpcConn.Bootstrap(ctx)
	return rpcConn, hubClient, err
}

// ConnectToService returns a client connected to the given service address using tcp or websockets.
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
func ConnectToService(
	fullUrl string, searchTimeSec int, clientCert *tls.Certificate, caCert *x509.Certificate) (conn net.Conn, err error) {

	// var tlsConn *tls.Conn
	const timeout = time.Second * 3
	var tlsConfig *tls.Config
	var clientCertList []tls.Certificate = nil
	var checkServerCert bool
	caCertPool := x509.NewCertPool()

	// use gateway discovery
	if fullUrl == "" {
		fullUrl = LocateHub("", searchTimeSec)
	}

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
		// Unix domain socket. TLS is not needed.
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

// ConnectToUDS creates a connection to a service using UDS
// using the convention that connection address = socketPath/service.socket
func ConnectToUDS(serviceName, socketFolder string) (net.Conn, error) {
	timeout := time.Second * 3
	socketPath := fmt.Sprintf("%s/%s.socket", socketFolder, serviceName)
	return net.DialTimeout("unix", socketPath, timeout)
}
