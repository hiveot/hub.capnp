// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
)

// UserCertsCapnpClient provides the POGS wrapper around the Capnp API
// This implements the IUserCerts interface
type UserCertsCapnpClient struct {
	capability hubapi.CapUserCerts
}

// CreateUserCert creates a CA signed certificate for mutual authentication by consumers
func (cl *UserCertsCapnpClient) CreateUserCert(
	ctx context.Context, clientID string, pubKeyPEM string, validityDays int) (
	certPEM string, caCertPEM string, err error) {

	method, release := cl.capability.CreateUserCert(ctx,
		func(params hubapi.CapUserCerts_createUserCert_Params) error {
			err2 := params.SetClientID(clientID)
			_ = params.SetPubKeyPEM(pubKeyPEM)
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
func (cl *UserCertsCapnpClient) Release() {
	cl.capability.Release()
}

// NewUserCertsCapnpClient returns a POGO wrapper for the capnp protocol
// This is for internal use. The capability has to be obtained using CertsCapnpClient.
func NewUserCertsCapnpClient(cap hubapi.CapUserCerts) *UserCertsCapnpClient {
	cl := &UserCertsCapnpClient{capability: cap}
	return cl
}
