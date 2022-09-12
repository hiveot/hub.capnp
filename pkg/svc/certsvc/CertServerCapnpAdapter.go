// capnproto adapter for handling rpc calls
package main

import (
	"context"
	"fmt"
	"net"

	capnp "capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/capnp/svc"

	"github.com/hiveot/hub/pkg/svc/certsvc/selfsigned"
	"github.com/hiveot/hub/pkg/svc/certsvc/service"
)

// CertServerCapnpAdapter implements the capnproto generated interface CertService_Server
type CertServerCapnpAdapter struct {
	srv service.ICertService
}

func (csca *CertServerCapnpAdapter) CreateClientCert(
	ctx context.Context, call svc.CertService_createClientCert) error {
	fmt.Println("CertServerCapnpAdapter.createClientCert")

	clientID, _ := call.Args().ClientID()
	pubKeyPEM, _ := call.Args().PubKeyPEM()
	certPEM, caCertPEM, err := csca.srv.CreateClientCert(clientID, pubKeyPEM)
	if err == nil {
		logrus.Infof("CertServerCapnpAdapter Created client cert for %s", clientID)
		res, err2 := call.AllocResults()
		res.SetCertPEM(certPEM)
		res.SetCaCertPEM(caCertPEM)
		err = err2
	}
	return err
}

func (csca *CertServerCapnpAdapter) CreateDeviceCert(
	context.Context, svc.CertService_createDeviceCert) error {
	return nil
}

func (csca *CertServerCapnpAdapter) CreateServiceCert(context.Context, svc.CertService_createServiceCert) error {
	return nil
}

// capnproto server
func ServeCertServiceCapnpAdapter(ctx context.Context,
	srv *selfsigned.SelfSignedCertService,
	lis net.Listener) error {

	main := svc.CertService_ServerToClient(&CertServerCapnpAdapter{
		srv: srv,
	})
	// Listen for calls
	for {
		rwc, _ := lis.Accept()
		go func() error {
			transport := rpc.NewStreamTransport(rwc)
			conn := rpc.NewConn(transport, &rpc.Options{
				BootstrapClient: capnp.Client(main.AddRef()),
			})
			defer conn.Close()
			// Wait for connection to abort.
			select {
			case <-conn.Done():
				return nil
			case <-ctx.Done():
				return conn.Close()
			}
		}()
	}
}
