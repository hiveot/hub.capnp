package listener

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"capnproto.org/go/capnp/v3/rpc/transport"
	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
	websocketcapnp "zenhack.net/go/websocket-capnp"
)

// ServeWSCB incoming websocket connections to a handler
func ServeWSCB(lis net.Listener, wsPath string, handler func(net.Conn, transport.Transport)) error {
	if handler == nil {
		panic("connection handler required")
	}
	router := http.NewServeMux()

	// upgrade incoming http connections to websocket
	// FIXME: /ws path should be configurable
	router.HandleFunc(wsPath, func(w http.ResponseWriter, req *http.Request) {
		logrus.Infof("From host=%s, method=%s", req.Host, req.Method)
		up := ws.HTTPUpgrader{}

		// from: codec, err := websocketcapnp.UpgradeHTTP(up, r, w)
		conn, bufRw, _, err := up.Upgrade(req, w)
		if err != nil {
			err = fmt.Errorf("error upgrading websocket connection: %w", err)
			logrus.Error(err)
			return
		}
		if n := bufRw.Reader.Buffered(); n > 0 {
			err = fmt.Errorf("TODO: support buffered data on hijacked connection (%v bytes buffered)", n)
			logrus.Error(err)
			return
		}
		if err := bufRw.Writer.Flush(); err != nil {
			err = fmt.Errorf("Flush(): %w", err)
			logrus.Error(err)
			return
		}
		// this new websocket server connection can be used to create a transport by
		// providing the capnp codecs.
		codec := websocketcapnp.NewCodec(conn, true)
		capTransport := transport.New(codec)
		handler(conn, capTransport)
	})
	err := http.Serve(lis, router)
	return err
}

// ServeWS serves a Cap'n Proto RPC over websockets.
//
// Serve will take ownership of bootstrap client and release it after the listener closes.
// Serve exits with the listener error if the listener is closed by the owner.
//
//		lis is the tcp or TLS socket listener
//	 wsPath is the websocket path to listen on. "/ws" is recommended
//		boot is the capnp client to handle the requests
//		onConnect is optional and called when a new connection is established
//		onDisconnect is optional and called when a connection closes
func ServeWS(serviceID string,
	lis net.Listener,
	wsPath string,
	boot capnp.Client,
	onConnect func(rpcConn *rpc.Conn),
	onDisconnect func(rpcConn *rpc.Conn),
) error {

	if !boot.IsValid() {
		err := errors.New("bootstrapClient for service " + serviceID + "is not valid")
		logrus.Error(err)
		return err
	}
	// Since we took ownership of the bootstrap client, release it after we're done.
	defer boot.Release()

	err := ServeWSCB(lis, wsPath, func(_ net.Conn, tp transport.Transport) {
		// the RPC connection takes ownership of the bootstrap interface and will release it when the connection
		// exits, so use AddRef when providing the bootstrap to avoid releasing the provided bootstrap client capability.
		opts := rpc.Options{
			BootstrapClient: boot.AddRef(),
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
		}()
		// ServeTransport(serviceID, tp, boot.AddRef(), onConnect, onDisconnect)
	})
	return err

}

// // ListenAndServeTLSWS wraps the given listener in a TLS websocket connection.
// // This takes ownership of bootstrapClient and release it on exit.
// //
// // lis is the incoming connection listener to use.
// // boot is the capnp bootstrap client.
// // serverCert is this server's TLS certificate, signed by the CA.
// // caCert is the CA certificate that has signed the server certificate
// func ListenAndServeTLSWS(ctx context.Context,
// 	serviceID string, lis net.Listener, boot capnp.Client,
// 	serverCert *tls.Certificate, caCert *x509.Certificate) (err error) {

// 	tlsLis := CreateTLSListener(lis, serverCert, caCert)

// 	// to close this listener, close the context
// 	go func() {
// 		<-ctx.Done()
// 		_ = tlsLis.Close()
// 	}()
// 	err = ServeWS(serviceID, tlsLis, boot, nil, nil)

// 	return err
// }
