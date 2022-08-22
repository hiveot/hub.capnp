package selfsigned

import (
	"context"
	"crypto/x509"

	"github.com/wostzone/wost-go/pkg/certsclient"
	"github.com/wostzone/wost.grpc/go/svc"
	"svc/certsvc/config"
)

// SelfSignedServer implements the svc.CertServiceServer interface
// This service creates certificates for use by services, devices (via idprov) and admin users.
// Note that this service does not support certificate revocation.
//   See also: https://www.imperialviolet.org/2014/04/19/revchecking.html
// Instead the issued certificates are short lived and must be renewed before they expire.
type SelfSignedServer struct {
	svc.UnimplementedCertServiceServer
	config config.CertSvcConfig
}

// CreateClientCert creates a CA signed certificate for mutual authentication by consumers
func (srv *SelfSignedServer) CreateClientCert(_ context.Context, args *svc.CreateClientCert_Args) (*svc.Cert_Res, error) {
	pubKey, err := certsclient.PublicKeyFromPEM(args.PubKeyPEM)
	if err != nil {
		return nil, err
	}

	cert, err := CreateClientCert(
		args.ClientID,
		certsclient.OUClient,
		pubKey,
		srv.config.CaCert,
		srv.config.CaKey,
		config.DefaultClientCertDurationDays)

	caCertPem := certsclient.X509CertToPEM(srv.config.CaCert)
	certPem := certsclient.X509CertToPEM(cert)
	res := &svc.Cert_Res{
		CertPEM:   certPem,
		CaCertPEM: caCertPem,
	}
	return res, nil
}

// CreateDeviceCert creates a CA signed certificate for mutual authentication by IoT devices
func (srv *SelfSignedServer) CreateDeviceCert(_ context.Context, args *svc.CreateClientCert_Args) (*svc.Cert_Res, error) {
	var res = &svc.Cert_Res{}
	var cert *x509.Certificate
	var err error
	pubKey, err := certsclient.PublicKeyFromPEM(args.PubKeyPEM)
	if err == nil {
		cert, err = CreateClientCert(
			args.ClientID,
			certsclient.OUIoTDevice,
			pubKey,
			srv.config.CaCert,
			srv.config.CaKey,
			config.DefaultDeviceCertDurationDays)

		caCertPem := certsclient.X509CertToPEM(srv.config.CaCert)
		certPem := certsclient.X509CertToPEM(cert)
		res.CertPEM = certPem
		res.CaCertPEM = caCertPem
	}
	return res, err
}

// CreateServiceCert creates a CA signed service certificate for mutual authentication between services
func (srv *SelfSignedServer) CreateServiceCert(_ context.Context, args *svc.CreateServiceCert_Args) (*svc.Cert_Res, error) {
	pubKey, err := certsclient.PublicKeyFromPEM(args.PubKeyPEM)
	if err != nil {

	}

	cert, err := CreateServiceCert(
		args.ServiceID,
		args.Names,
		pubKey,
		srv.config.CaCert,
		srv.config.CaKey,
		config.DefaultServiceCertDurationDays,
	)

	caCertPem := certsclient.X509CertToPEM(srv.config.CaCert)
	certPem := certsclient.X509CertToPEM(cert)
	res := &svc.Cert_Res{
		CertPEM:   certPem,
		CaCertPEM: caCertPem,
	}
	return res, nil
}

func NewSelfSignedServer(config config.CertSvcConfig) *SelfSignedServer {
	service := &SelfSignedServer{
		config: config,
	}
	return service
}
