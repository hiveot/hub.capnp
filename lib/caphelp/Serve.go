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

	"github.com/hiveot/hub/lib/listener"
)

// Serve serves a Cap'n Proto RPC to incoming connections.
//
// Serve will take ownership of bootstrap client and release it after the listener closes.
//
// Serve exits with the listener error if the listener is closed by the owner.
func Serve(serviceID string, lis net.Listener, boot capnp.Client, onConnection func(rpcConn *rpc.Conn)) error {

	if !boot.IsValid() {
		err := errors.New("BootstrapClient is not valid")
		return err
	}
	// Since we took ownership of the bootstrap client, release it after we're done.
	defer boot.Release()
	for {
		// Accept incoming connections
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		connID := GetConnectionID(conn)

		logrus.Infof("Incoming connection to '%s' from remote client: %s. ID=%s",
			serviceID, conn.RemoteAddr().String(), connID)

		// the RPC connection takes ownership of the bootstrap interface and will release it when the connection
		// exits, so use AddRef to avoid releasing the provided bootstrap client capability.
		opts := rpc.Options{
			BootstrapClient: boot.AddRef(),
		}
		// For each new incoming connection, create a new RPC transport connection that will serve incoming RPC requests
		transport := rpc.NewStreamTransport(conn)
		rpcConn := rpc.NewConn(transport, &opts)

		// the RPC connection is now established and we can retrieve the remote bootstrap client
		// if the service is using it.
		if onConnection != nil {
			onConnection(rpcConn)
		}
		go func() {
			logrus.Infof("Connection to '%s' with ID='%s' established", serviceID, connID)
			<-rpcConn.Done()
			logrus.Infof("Connection to '%s' with ID='%s' closed", serviceID, connID)
		}()
	}
}

//// ServeTLS serves a Cap'n Proto RPC to incoming connections using TLS
////
//// Serve will take ownership of bootstrapClient and release it after the listener closes.
//// Serve exits with the listener error if the listener is closed by the owner.
////
//// bootstrapClient is the service that will handle Cap'n Proto RPC requests. It is
//// obtained with Xyz_ServerToClient method that implements the capability defined
//// in the capnp schema for interface Xyz.
////
//// serverCert is this server's TLS certificate, signed by the CA.
//// caCert is the CA certificate that has signed the server certificate
//func ServeTLS(lis net.Listener, boot capnp.Client,
//	serverCert *tls.Certificate, caCert *x509.Certificate) error {
//
//	if !boot.IsValid() {
//		err := errors.New("bootstrap client is not valid")
//		return err
//	}
//
//	caCertPool := x509.NewCertPool()
//	caCertPool.AddCert(caCert)
//	tlsConfig := &tls.Config{
//		Certificates:       []tls.Certificate{*serverCert},
//		ClientAuth:         tls.VerifyClientCertIfGiven,
//		ClientCAs:          caCertPool,
//		MinVersion:         tls.VersionTLS12,
//		InsecureSkipVerify: false,
//	}
//
//	// Accept incoming connections
//	defer boot.Release()
//	for {
//		conn, err := lis.Accept()
//		if err != nil {
//			// Since we took ownership of the bootstrap client, release it after we're done.
//			boot.Release()
//			return err
//		}
//		connID := GetConnectionID(conn)
//		logrus.Infof("New connection from remote client: %s. ID=%s",
//			conn.RemoteAddr().String(), connID)
//
//		// turn the transport into a TLS connection
//		tlsConn := tls.Server(conn, tlsConfig)
//
//		// the RPC connection takes ownership of the bootstrap interface and will release it when the connection
//		// exits, so use AddRef to avoid releasing the provided bootstrap client capability.
//		opts := rpc.Options{
//			BootstrapClient: boot.AddRef(),
//		}
//		// For each new incoming connection, create a new RPC transport connection that
//		// will serve incoming RPC requests
//		transport := rpc.NewStreamTransport(tlsConn)
//		_ = rpc.NewConn(transport, &opts)
//	}
//}

// ListenAndServe opens a listener on the given address and serves a Cap'n Proto RPC to incoming connections
//
// network and address are passed to net.Listen. Use network "unix" for Unix Domain Sockets
// and "tcp" for regular TCP IP4 or IP6 connections.
//
// ListenAndServe will take ownership of bootstrapClient and release it on exit.
func ListenAndServe(ctx context.Context, serviceID string, network, addr string, bootstrapClient capnp.Client) error {

	listener, err := net.Listen(network, addr)

	if err == nil {
		// to close this listener, close the context
		go func() {
			<-ctx.Done()
			_ = listener.Close()
		}()
		err = Serve(serviceID, listener, bootstrapClient, nil)
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

	tlsLis := listener.CreateTLSListener(lis, serverCert, caCert)

	// to close this listener, close the context
	go func() {
		<-ctx.Done()
		_ = tlsLis.Close()
	}()
	err = Serve(serviceID, tlsLis, boot, nil)

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
