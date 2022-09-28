package selfsigned

import (
	"crypto/ecdsa"
	"crypto/x509"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/certs"
)

// SelfSignedCertsService creates certificates for use by services, devices and admin users.
//
// This implements the ICertsService interface
//
// Note that this service does not support certificate revocation.
//   See also: https://www.imperialviolet.org/2014/04/19/revchecking.html
// Issued certificates are short-lived and must be renewed before they expire.
type SelfSignedCertsService struct {
	deviceCertsService  certs.IDeviceCerts
	serviceCertsService certs.IServiceCerts
	userCertsService    certs.IUserCerts
	verifyCertsService  certs.IVerifyCerts
}

// CapDeviceCerts provides the capability to manage device certificates
func (srv *SelfSignedCertsService) CapDeviceCerts() certs.IDeviceCerts {
	return srv.deviceCertsService
}

// CapServiceCerts provides the capability to manage service certificates
func (srv *SelfSignedCertsService) CapServiceCerts() certs.IServiceCerts {
	return srv.serviceCertsService
}

// CapUserCerts provides the capability to manage user certificates
func (srv *SelfSignedCertsService) CapUserCerts() certs.IUserCerts {
	return srv.userCertsService
}

// CapVerifyCerts provides the capability to verify certificates
func (srv *SelfSignedCertsService) CapVerifyCerts() certs.IVerifyCerts {
	return srv.verifyCertsService
}

// Release the provided capabilities
// nothing to do here
func (srv *SelfSignedCertsService) Release() {
}

// NewSelfSignedCertsService returns a new instance of the selfsigned certificate service
//  caCert is the CA certificate used to created certificates
//  caKey is the CA private key used to created certificates
func NewSelfSignedCertsService(caCert *x509.Certificate, caKey *ecdsa.PrivateKey) *SelfSignedCertsService {
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)

	// Use one service instance per capability.
	// This does open the door to creating an instance per client session with embedded constraints,
	// although this is not used at the moment.
	service := &SelfSignedCertsService{
		deviceCertsService:  NewDeviceCertsService(caCert, caKey),
		serviceCertsService: NewServiceCertsService(caCert, caKey),
		userCertsService:    NewUserCertsService(caCert, caKey),
		verifyCertsService:  NewVerifyCertsService(caCert),
	}
	if caCert == nil || caKey == nil || caCert.PublicKey == nil {
		logrus.Panic("Missing CA certificate or key")
	}

	return service
}
