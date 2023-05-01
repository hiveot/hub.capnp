package hubclient

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"capnproto.org/go/capnp/v3/rpc/transport"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
	"net"
	"net/url"
	"time"
	websocketcapnp "zenhack.net/go/websocket-capnp"
)

// ConnectWithCapnpWebsockets returns a capnp connection to the given endpoint using websockets.
//
// This accepts a client certificate for client authentication and a server CA certificate
// to verify the server connection. If neither client nor CA certificate are provided TLS is not used.
// If the url schema is 'unix' then a local UDS is used and certificates are ignored.
//
// Note that when connecting to websockets, capnp needs a special media encoder to talk to the
// server. This is automatically applied when the URL starts with ws:// or wss://.
//
// To close the connection, invoke Release on the client.
//
//	fullUrl supports both tcp and wss for websocket connections: wss://server:port/ws
//
//	clientCert is the client certificate to authenticate with. Use nil to not use client authentication
//	caCert is the CA certificate used to verify the server authenticity. Use nil if server auth is not yet established.
//
// Returns a capnp client using the websocket codec
func ConnectWithCapnpWebsockets(
	fullUrl string, clientCert *tls.Certificate, caCert *x509.Certificate) (cap capnp.Client, err error) {

	// var tlsConn *tls.Conn
	var conn net.Conn
	const timeout = time.Second * 3
	var tlsConfig *tls.Config
	var clientCertList []tls.Certificate = nil
	var checkServerCert bool
	caCertPool := x509.NewCertPool()

	// use gateway discovery
	if fullUrl == "" {
		return cap, errors.New("missing service URL")
		// this can add 0.6 MB of code so lets leave it up to the user whether this is needed
		//fullUrl = LocateHub("", searchTimeSec)
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
				return cap, err
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
	// default url with wss
	u, err := url.Parse(fullUrl)
	if err != nil {
		u = &url.URL{Scheme: "wss", Host: fullUrl}
		fullUrl = "wss://" + fullUrl
	}
	if u.Scheme == "ws" {
		// websocket no TLS
		dialer := ws.Dialer{Timeout: time.Second * 3}
		// TODO: check if a buffer is returned and return it to the pool if not nil
		conn, _, _, err = dialer.Dial(context.Background(), fullUrl)
	} else {
		// websocket with TLS - falls back to no TLS if tlsConfig is nil
		// TODO: check if a buffer is returned and return it to the pool if not nil
		dialer := ws.Dialer{TLSConfig: tlsConfig, Timeout: time.Second * 3}
		conn, _, _, err = dialer.Dial(context.Background(), fullUrl)
	}

	if err != nil {
		err = errors.New("Unable to connect to " + fullUrl + ". Error: " + err.Error())
		logrus.Error(err)
	} else {
		//logrus.Infof("connected to '%s'", fullUrl)
		codec := websocketcapnp.NewCodec(conn, false)
		tp := transport.New(codec)
		rpcConn := rpc.NewConn(tp, nil)
		cap = rpcConn.Bootstrap(context.Background())
	}
	return cap, err
}
