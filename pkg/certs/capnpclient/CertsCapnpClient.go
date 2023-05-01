// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/certs"
)

// CertsCapnpClient provides the POGS wrapper around the capnp CapCerts API
// This implements the ICerts interface
//
// Getting the capability should somehow be tied to the user's permissions. How this auth
// aspect is handled is tbd.
type CertsCapnpClient struct {
	connection *rpc.Conn // connection to capnp server
	capability hubapi.CapCerts
}

// CapDeviceCerts returns the capability to create device certificates
func (cl *CertsCapnpClient) CapDeviceCerts(ctx context.Context, clientID string) (certs.IDeviceCerts, error) {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapDeviceCerts(ctx,
		func(params hubapi.CapCerts_capDeviceCerts_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	defer release()
	capability := getCap.Cap().AddRef()
	newCap := NewDeviceCertsCapnpClient(capability)
	return newCap, nil
}

// CapServiceCerts returns the capability to create service certificates
func (cl *CertsCapnpClient) CapServiceCerts(ctx context.Context, clientID string) (certs.IServiceCerts, error) {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapServiceCerts(ctx,
		func(params hubapi.CapCerts_capServiceCerts_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	defer release()
	capability := getCap.Cap().AddRef()
	newCap := NewServiceCertsCapnpClient(capability)
	return newCap, nil
}

// CapUserCerts returns the capability to create user certificates
func (cl *CertsCapnpClient) CapUserCerts(ctx context.Context, clientID string) (certs.IUserCerts, error) {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapUserCerts(ctx,
		func(params hubapi.CapCerts_capUserCerts_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	defer release()
	capability := getCap.Cap().AddRef()
	newCap := NewUserCertsCapnpClient(capability)
	return newCap, nil
}

// CapVerifyCerts returns the capability to verify certificates
func (cl *CertsCapnpClient) CapVerifyCerts(ctx context.Context, clientID string) (certs.IVerifyCerts, error) {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapVerifyCerts(ctx,
		func(params hubapi.CapCerts_capVerifyCerts_Params) error {
			err2 := params.SetClientID(clientID)
			return err2

		})
	defer release()
	capability := getCap.Cap().AddRef()
	newCap := NewVerifyCertsCapnpClient(capability)
	return newCap, nil
}

// Release the provided capabilities after use and release resources
func (cl *CertsCapnpClient) Release() {
	cl.capability.Release()
}

// NewCertsCapnpClient returns a capability to create certificates //
func NewCertsCapnpClient(capClient capnp.Client) *CertsCapnpClient {
	capability := hubapi.CapCerts(capClient)

	cl := &CertsCapnpClient{
		connection: nil,
		capability: capability,
	}
	return cl
}
