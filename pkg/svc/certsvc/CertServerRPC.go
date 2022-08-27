package main

import (
	"context"

	"github.com/wostzone/hub/pkg/svc/certsvc/selfsigned"
	"github.com/wostzone/hub/pkg/svc/certsvc/service"
	"github.com/wostzone/wost.grpc/go/svc"
)

// CertServerRPC is the gRPC service interface for the self-signed certificate service
type CertServerRPC struct {
	svc.UnimplementedCertServiceServer
	srv service.ICertService
}

func (rpc *CertServerRPC) CreateClientCert(_ context.Context, args *svc.CreateClientCert_Args) (*svc.Cert_Res, error) {
	certPem, caCertPem, err := rpc.srv.CreateClientCert(args.ClientID, args.PubKeyPEM)

	res := &svc.Cert_Res{
		CertPEM:   certPem,
		CaCertPEM: caCertPem,
	}
	return res, err
}

func (rpc *CertServerRPC) CreateDeviceCert(_ context.Context, args *svc.CreateClientCert_Args) (*svc.Cert_Res, error) {
	certPem, caCertPem, err := rpc.srv.CreateDeviceCert(args.ClientID, args.PubKeyPEM)

	res := &svc.Cert_Res{
		CertPEM:   certPem,
		CaCertPEM: caCertPem,
	}
	return res, err
}

func (rpc *CertServerRPC) CreateServiceCert(_ context.Context, args *svc.CreateServiceCert_Args) (*svc.Cert_Res, error) {
	certPem, caCertPem, err := rpc.srv.CreateServiceCert(args.ServiceID, args.PubKeyPEM, args.Names)

	res := &svc.Cert_Res{
		CertPEM:   certPem,
		CaCertPEM: caCertPem,
	}
	return res, err
}

func NewCertServerRPC(srv *selfsigned.SelfSignedCertService) *CertServerRPC {
	rpcServer := &CertServerRPC{srv: srv}
	return rpcServer
}
