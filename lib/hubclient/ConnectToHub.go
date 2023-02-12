package hubclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"strings"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"capnproto.org/go/capnp/v3/rpc/transport"
	websocketcapnp "zenhack.net/go/websocket-capnp"

	"github.com/hiveot/hub/pkg/resolver"
)

// ConnectToHub using TLS and discovery.
//
//	 This connects to the resolver socket if available, or to the gateway if not.
//	 DNS-SD of the gateway is planned.
//
//		fullURL is optional connection endpoint. Use "" to auto discover the resolver or gateway.
func ConnectToHub(fullUrl string, clientCert *tls.Certificate, caCert *x509.Certificate) (net.Conn, error) {

	// TODO: use discovery
	if fullUrl == "" {
		fullUrl = "unix://" + resolver.DefaultResolverPath
	}

	return CreateClientConnection(fullUrl, clientCert, caCert)
}

// ConnectToHubClient returns the connection and capnp client of the gateway or resolver service.
//
// This client is special in that a request for a capability is dynamically forwarded to the actual service.
// If the client does not have the proper authentication type then the capability is not available and this fails
// with 'unimplemented'.
//
// Note that when connecting without client certificate to the gateway, Login must be called to authenticate.
//
// This auto-discovers the gateway or default to 127.0.0.1:8883
func ConnectToHubClient(
	fullUrl string, clientCert *tls.Certificate, caCert *x509.Certificate) (
	rpcCon *rpc.Conn, cap capnp.Client, err error) {

	var tp transport.Transport

	conn, err := ConnectToHub(fullUrl, clientCert, caCert)
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
