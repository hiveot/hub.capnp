package caphelp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"
)

// Serve serves a Cap'n Proto RPC to incoming connections.
//
// Serve will take ownership of bootstrapClient and release it after the listener closes.
//
// Serve exits with the listener error if the listener is closed by the owner.
func Serve(lis net.Listener, bootstrapClient capnp.Client) error {
	if !bootstrapClient.IsValid() {
		err := errors.New("BootstrapClient is not valid")
		return err
	}
	// Accept incoming connections
	defer bootstrapClient.Release()
	for {
		conn, err := lis.Accept()
		if err != nil {
			// Since we took ownership of the bootstrap client, release it after we're done.
			if !bootstrapClient.IsValid() {
				err = errors.New("the bootstrap client was already released")
			}
			return err
		}
		connID := getConnectionID(conn)

		logrus.Infof("New connection from remote client: %s. ID=%s",
			conn.RemoteAddr().String(), connID)

		// the RPC connection takes ownership of the bootstrap interface and will release it when the connection
		// exits, so use AddRef to avoid releasing the provided bootstrap client capability.
		opts := rpc.Options{
			BootstrapClient: bootstrapClient.AddRef(),
		}

		// For each new incoming connection, create a new RPC transport connection that will serve incoming RPC requests
		// rpc.Options will contain the bootstrap capability
		go func() {
			transport := rpc.NewStreamTransport(conn)
			conn := rpc.NewConn(transport, &opts)
			<-conn.Done()
			// Remote client connection closed
		}()
	}
}

// ServeTLS serves a Cap'n Proto RPC to incoming connections using TLS
//
// Serve will take ownership of bootstrapClient and release it after the listener closes.
// Serve exits with the listener error if the listener is closed by the owner.
//
// bootstrapClient is the service that will handle Cap'n Proto RPC requests. It is
// obtained with Xyz_ServerToClient method that implements the capability defined
// in the capnp schema for interface Xyz.
//
// serverCert is this server's TLS certificate, signed by the CA.
// caCert is the CA certificate that has signed the server certificate
func ServeTLS(lis net.Listener, bootstrapClient capnp.Client,
	serverCert *tls.Certificate, caCert *x509.Certificate) error {

	if !bootstrapClient.IsValid() {
		err := errors.New("BootstrapClient is not valid")
		return err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{*serverCert},
		ClientAuth:         tls.VerifyClientCertIfGiven,
		ClientCAs:          caCertPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	// Accept incoming connections
	defer bootstrapClient.Release()
	for {
		conn, err := lis.Accept()
		if err != nil {
			// Since we took ownership of the bootstrap client, release it after we're done.
			if !bootstrapClient.IsValid() {
				err = errors.New("the bootstrap client was already released")
			}
			return err
		}
		connID := getConnectionID(conn)
		logrus.Infof("New connection from remote client: %s. ID=%s",
			conn.RemoteAddr().String(), connID)

		// turn the transport into a TLS connection
		tlsConn := tls.Server(conn, tlsConfig)
		// For each new incoming connection, create a new RPC transport connection that
		// will serve incoming RPC requests
		transport := rpc.NewStreamTransport(tlsConn)

		// the RPC connection takes ownership of the bootstrap interface and will release it when the connection
		// exits, so use AddRef to avoid releasing the provided bootstrap client capability.
		opts := rpc.Options{
			BootstrapClient: bootstrapClient.AddRef(),
		}

		go func() {
			rpcConn := rpc.NewConn(transport, &opts)
			<-rpcConn.Done()
			// Remote client connection closed
		}()
	}
}

// ListenAndServe opens a listener on the given address and serves a Cap'n Proto RPC to incoming connections
//
// network and address are passed to net.Listen. Use network "unix" for Unix Domain Sockets
// and "tcp" for regular TCP IP4 or IP6 connections.
//
// ListenAndServe will take ownership of bootstrapClient and release it on exit.
func ListenAndServe(ctx context.Context, network, addr string, bootstrapClient capnp.Client) error {

	listener, err := net.Listen(network, addr)

	if err == nil {
		// to close this listener, close the context
		go func() {
			<-ctx.Done()
			_ = listener.Close()
		}()
		err = Serve(listener, bootstrapClient)
	}
	return err
}

// ListenAndServeTLS opens a listener for TLS connections on the given address and serves
// a Cap'n Proto RPC to incoming connections.
//
// network and address are passed to net.Listen. Use network "unix" for Unix Domain Sockets
// and "tcp" for regular TCP IP4 or IP6 connections.
//
// ListenAndServe will take ownership of bootstrapClient and release it on exit.
//
// serverCert is this server's TLS certificate, signed by the CA.
// caCert is the CA certificate that has signed the server certificate
func ListenAndServeTLS(ctx context.Context,
	network, addr string, bootstrapClient capnp.Client,
	serverCert *tls.Certificate, caCert *x509.Certificate) error {

	listener, err := net.Listen(network, addr)

	if err == nil {
		// to close this listener, close the context
		go func() {
			<-ctx.Done()
			_ = listener.Close()
		}()
		err = ServeTLS(listener, bootstrapClient, serverCert, caCert)
	}
	return err
}

// getConnectionID returns the ID of the unix domain or TCP socket connection.
// used to pair incoming and closing connections in the logs
// This returns 0 if the connection is not unix or tcp, or is closed.
func getConnectionID(conn net.Conn) string {
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
