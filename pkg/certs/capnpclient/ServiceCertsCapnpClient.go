// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
)

// ServiceCertsCapnpClient provides the POGS wrapper around the Capnp API
// This implements the IServiceCerts interface
type ServiceCertsCapnpClient struct {
	capability hubapi.CapServiceCerts
}

func (cl *ServiceCertsCapnpClient) CreateServiceCert(
	ctx context.Context, serviceID string, pubKeyPEM string, names []string, validityDays int) (
	certPEM string, caCertPEM string, err error) {

	method, release := cl.capability.CreateServiceCert(ctx,
		func(params hubapi.CapServiceCerts_createServiceCert_Params) error {
			err2 := params.SetServiceID(serviceID)
			_ = params.SetPubKeyPEM(pubKeyPEM)
			if names != nil {
				_ = params.SetNames(caphelp.MarshalStringList(names))
			}
			params.SetValidityDays(int32(validityDays))
			return err2
		})
	defer release()
	resp2, err := method.Struct()
	if err == nil {
		certPEM, err = resp2.CertPEM()
		caCertPEM, _ = resp2.CaCertPEM()
	}
	return certPEM, caCertPEM, err
}

// Release the provided capabilities after use and release resources
func (cl *ServiceCertsCapnpClient) Release() {
	cl.capability.Release()
}

// NewServiceCertsCapnpClient returns a capability to create certificates using the capnp protocol
// This is for internal use. The capability has to be obtained using CertsCapnpClient.
func NewServiceCertsCapnpClient(cap hubapi.CapServiceCerts) *ServiceCertsCapnpClient {
	cl := &ServiceCertsCapnpClient{capability: cap}
	return cl
}
