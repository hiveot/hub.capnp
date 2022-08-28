package main

import (
	"context"
	"io"
	"time"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/capnp/svc"
	"github.com/hiveot/hub/pkg/svc/certsvc/selfsigned"
	"github.com/hiveot/hub/pkg/svc/certsvc/service"
)

// CertServerCapnpAdapter implements the capnproto generated interface CertService_Server
type CertServerCapnpAdapter struct {
	srv service.ICertService
}

func (csca *CertServerCapnpAdapter) CreateClientCert(context.Context, svc.CertService_createClientCert) error {
	return nil
}

func (csca *CertServerCapnpAdapter) CreateDeviceCert(context.Context, svc.CertService_createDeviceCert) error {
	return nil
}
func (csca *CertServerCapnpAdapter) CreateServiceCert(context.Context, svc.CertService_createServiceCert) error {
	return nil
}

//
func ServeCertService(srv *selfsigned.SelfSignedCertService, rwc io.ReadWriteCloser) error {
	main := svc.CertService_ServerToClient(&CertServerCapnpAdapter{
		srv: srv,
	})
	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	// Listen for calls, using the logger as the bootstrap interface.
	conn := rpc.NewConn(rpc.NewStreamTransport(rwc), &rpc.Options{
		BootstrapClient: main.Client,
	})
	defer conn.Close()

	// Wait for connection to abort.
	select {
	case <-conn.Done():
		return nil
	case <-ctx.Done():
		return conn.Close()
	}
}
