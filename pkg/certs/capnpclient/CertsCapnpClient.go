// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
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
	ctx        context.Context
}

// CapDeviceCerts returns the capability to create device certificates
func (cl *CertsCapnpClient) CapDeviceCerts() certs.IDeviceCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapDeviceCerts(cl.ctx, nil)
	defer release()
	capability := getCap.Cap().AddRef()
	return NewDeviceCertsCapnpClient(capability)
}

// CapServiceCerts returns the capability to create service certificates
func (cl *CertsCapnpClient) CapServiceCerts() certs.IServiceCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapServiceCerts(cl.ctx, nil)
	defer release()
	capability := getCap.Cap().AddRef()
	return NewServiceCertsCapnpClient(capability)
}

// CapUserCerts returns the capability to create user certificates
func (cl *CertsCapnpClient) CapUserCerts() certs.IUserCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapUserCerts(cl.ctx, nil)
	defer release()
	capability := getCap.Cap().AddRef()
	return NewUserCertsCapnpClient(capability)
}

// CapVerifyCerts returns the capability to verify certificates
func (cl *CertsCapnpClient) CapVerifyCerts() certs.IVerifyCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapVerifyCerts(cl.ctx, nil)
	defer release()
	capability := getCap.Cap().AddRef()
	return NewVerifyCertsCapnpClient(capability)
}

// Release the provided capabilities after use and release resources
func (cl *CertsCapnpClient) Release() {
	cl.capability.Release()
}

// NewCertServiceCapnpClient returns a capability to create certificates using the capnp protocol
// Intended for bootstrapping the capability chain
//  ctx is the context for retrieving capabilities
func NewCertServiceCapnpClient(ctx context.Context, conn net.Conn) (*CertsCapnpClient, error) {
	transport := rpc.NewStreamTransport(conn)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapCerts(rpcConn.Bootstrap(ctx))

	cl := &CertsCapnpClient{
		connection: rpcConn,
		capability: capability,
		ctx:        ctx,
	}
	return cl, nil
}
