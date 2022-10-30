// Package client that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// VerifyCertsCapnpClient provides the POGS wrapper around the Capnp API
// This implements the IVerifyCerts interface
type VerifyCertsCapnpClient struct {
	capability hubapi.CapVerifyCerts
}

// Release the provided capabilities after use and release resources
func (cl *VerifyCertsCapnpClient) Release() {
	cl.capability.Release()
}

// VerifyCert verifies is the given certificate is valid
func (cl *VerifyCertsCapnpClient) VerifyCert(
	ctx context.Context, clientID string, certPEM string) (err error) {

	method, release := cl.capability.VerifyCert(ctx,
		func(params hubapi.CapVerifyCerts_verifyCert_Params) error {
			err2 := params.SetClientID(clientID)
			_ = params.SetCertPEM(certPEM)
			return err2
		})
	defer release()
	_, err = method.Struct()
	return err

}

// NewVerifyCertsCapnpClient returns a capability to verify certificates using the capnp protocol
// This is for internal use. The capability has to be obtained using CertsCapnpClient.
func NewVerifyCertsCapnpClient(cap hubapi.CapVerifyCerts) *VerifyCertsCapnpClient {
	cl := &VerifyCertsCapnpClient{capability: cap}
	return cl
}
