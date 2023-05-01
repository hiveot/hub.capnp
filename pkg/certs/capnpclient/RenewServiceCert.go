package capnpclient

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
	"github.com/sirupsen/logrus"
)

// RenewServiceCert is a helper function to renew a service's server certificate.
//
// This returns the service certificate signed by the CA, and the certificate of
// the CA that signed the service cert. This panics if the certs service is not
// reachable.
//
// This requires access to the cert service UDS socket.
//
//	serviceID is the instance ID of the service used as the CN on the certificate
//	ipAddr ip address the service is listening on or "" for outbound IP and localhost
//	keys is the private key of the service
//	socketFolder is the location of the certs service UDS socket
func RenewServiceCert(
	serviceID string, ipAddr string, keys *ecdsa.PrivateKey, socketFolder string) (
	svcCert *tls.Certificate, caCert *x509.Certificate, err error) {

	var capServiceCert certs.IServiceCerts
	if ipAddr == "" {
		ip := listener.GetOutboundIP("")
		ipAddr = ip.String()
	}
	// the service certificate is valid for the localhost address and ip address
	names := []string{"127.0.0.1", ipAddr}
	pubKeyPEM, err := certsclient.PublicKeyToPEM(&keys.PublicKey)
	if err != nil {
		logrus.Errorf("invalid public key: %s", err)
		return nil, nil, err
	}

	ctx := context.Background()
	cap, err := hubclient.ConnectWithCapnpUDS(certs.ServiceName, socketFolder)
	if err != nil {
		logrus.Errorf("unable to connect to certs service: %s. Workaround with local instance", err)
		// FIXME: workaround or panic?
		capServiceCert = selfsigned.NewServiceCertsService(caCert, nil)
		return nil, nil, err
	} else {
		cs := NewCertsCapnpClient(cap)
		capServiceCert, err = cs.CapServiceCerts(ctx, hubapi.AuthTypeService)
		_ = err
	}
	svcPEM, caPEM, err := capServiceCert.CreateServiceCert(ctx, serviceID, pubKeyPEM, names, 0)
	if err != nil {
		logrus.Errorf("unable to create a service certificate: %s", err)
		return nil, nil, err
	}
	caCert, _ = certsclient.X509CertFromPEM(caPEM)
	privKeyPEM, _ := certsclient.PrivateKeyToPEM(keys)
	newSvcCert, err := tls.X509KeyPair([]byte(svcPEM), []byte(privKeyPEM))
	return &newSvcCert, caCert, err
}
