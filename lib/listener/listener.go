package listener

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"capnproto.org/go/capnp/v3/rpc/transport"
)

// ServeCB incoming connections to a handler
func ServeCB(lis net.Listener, handler func(net.Conn, transport.Transport)) error {
	for {
		// Accept incoming connections
		conn, err := lis.Accept()
		if err != nil {
			return err
		}

		// Accept incoming connections
		//connID := GetConnectionID(conn)

		//logrus.Infof("from: %s. ID=%s", conn.RemoteAddr().String(), connID)

		// Each connection gets a new RPC transport that will serve incoming RPC requests
		tp := rpc.NewStreamTransport(conn)

		handler(conn, tp)
	}
}

// Serve serves a Cap'n Proto RPC to incoming connections.
//
// Serve will take ownership of bootstrap client and release it after the listener closes.
//
// Serve exits with the listener error if the listener is closed by the owner.
func Serve(serviceID string,
	lis net.Listener,
	boot capnp.Client,
	onConnect func(rpcConn *rpc.Conn),
	onDisconnect func(rpcConn *rpc.Conn),
) error {

	if !boot.IsValid() {
		err := errors.New("BootstrapClient is not valid")
		return err
	}
	// Since we took ownership of the bootstrap client, release it after we're done.
	defer boot.Release()
	err := ServeCB(lis, func(c net.Conn, tp transport.Transport) {
		// ServeTransport(serviceID, tp, boot.AddRef(), onConnect, onDisconnect)

		// logrus.Infof("incoming connection to %s", c.LocalAddr())
		// the RPC connection takes ownership of the bootstrap interface and will release it when the connection
		// exits, so use AddRef when providing the bootstrap to avoid releasing the provided bootstrap client capability.
		opts := rpc.Options{
			BootstrapClient: boot.AddRef(),
		}
		// attach the client certificate of the connection to the bootstrap client
		clientCert, clientID, ou, err := GetPeerCert(c)
		if err != nil {
			opts.BootstrapClient.State().Metadata.Put("peerCert", clientCert)
			opts.BootstrapClient.State().Metadata.Put("clientID", clientID)
			opts.BootstrapClient.State().Metadata.Put("ou", ou)
		}
		// Each connection gets a new RPC transport that will serve incoming RPC requests
		// transport := rpc.NewStreamTransport(conn)
		rpcConn := rpc.NewConn(tp, &opts)

		go func() {
			// logrus.Infof("Connection to '%s' with ID='%s' established", serviceID, connID)
			// the RPC connection is now established and we can retrieve the remote bootstrap client
			// if the service is using it.
			if onConnect != nil {
				onConnect(rpcConn)
			}

			<-rpcConn.Done()

			if onDisconnect != nil {
				onDisconnect(rpcConn)
			}
			// logrus.Infof("Connection to '%s' with ID='%s' closed", serviceID, connID)
		}()

	})
	return err
}

// ListenAndServe opens a listener on the given address and serves a Cap'n Proto RPC to incoming connections
//
// network and address are passed to net.Listen. Use network "unix" for Unix Domain Sockets
// and "tcp" for regular TCP IP4 or IP6 connections.
//
// ListenAndServe will take ownership of bootstrapClient and release it on exit.
func ListenAndServe(ctx context.Context,
	serviceID string,
	network, addr string,
	bootstrapClient capnp.Client) error {

	listener, err := net.Listen(network, addr)

	if err == nil {
		// to close this listener, close the context
		go func() {
			<-ctx.Done()
			_ = listener.Close()
		}()
		err = Serve(serviceID, listener, bootstrapClient, nil, nil)
	}
	return err
}

// ListenAndServeTLS wraps the given listener in a TLS connections.
// This takes ownership of bootstrapClient and release it on exit.
//
// lis is the incoming connection listener to use.
// boot is the capnp bootstrap client.
// serverCert is this server's TLS certificate, signed by the CA.
// caCert is the CA certificate that has signed the server certificate
func ListenAndServeTLS(ctx context.Context,
	serviceID string, lis net.Listener, boot capnp.Client,
	serverCert *tls.Certificate, caCert *x509.Certificate) (err error) {

	tlsLis := CreateTLSListener(lis, serverCert, caCert)

	// to close this listener, close the context
	go func() {
		<-ctx.Done()
		_ = tlsLis.Close()
	}()
	err = Serve(serviceID, tlsLis, boot, nil, nil)

	return err
}

// GetConnectionID returns the ID of the unix domain or TCP socket connection.
// used to pair incoming and closing connections in the logs
// This returns 0 if the connection is not unix or tcp, or is closed.
func GetConnectionID(conn net.Conn) string {
	udc, found := conn.(*net.UnixConn)
	if found {
		fd, _ := udc.File()
		fdName := fd.Name()
		fdfd := fd.Fd()
		idText := fmt.Sprintf("%s [%d]", fdName, fdfd)
		return idText
	}
	tcp, found := conn.(*net.TCPConn)
	if found {
		ra := tcp.RemoteAddr()
		fd, _ := tcp.File()
		fdfd := fd.Fd()
		idText := fmt.Sprintf("%s [%d]", ra, fdfd)
		return idText
	}
	return "closed"
}
