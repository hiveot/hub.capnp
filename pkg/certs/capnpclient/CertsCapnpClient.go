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
}

// CapDeviceCerts returns the capability to create device certificates
func (cl *CertsCapnpClient) CapDeviceCerts(ctx context.Context) certs.IDeviceCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapDeviceCerts(ctx, nil)
	defer release()
	capability := getCap.Cap().AddRef()
	return NewDeviceCertsCapnpClient(capability)
}

// CapServiceCerts returns the capability to create service certificates
func (cl *CertsCapnpClient) CapServiceCerts(ctx context.Context) certs.IServiceCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapServiceCerts(ctx, nil)
	defer release()
	capability := getCap.Cap().AddRef()
	return NewServiceCertsCapnpClient(capability)
}

// CapUserCerts returns the capability to create user certificates
func (cl *CertsCapnpClient) CapUserCerts(ctx context.Context) certs.IUserCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapUserCerts(ctx, nil)
	defer release()
	capability := getCap.Cap().AddRef()
	return NewUserCertsCapnpClient(capability)
}

// CapVerifyCerts returns the capability to verify certificates
func (cl *CertsCapnpClient) CapVerifyCerts(ctx context.Context) certs.IVerifyCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.CapVerifyCerts(ctx, nil)
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
//
//	ctx is the context for retrieving capabilities
func NewCertServiceCapnpClient(conn net.Conn) (*CertsCapnpClient, error) {
	ctx := context.Background()
	transport := rpc.NewStreamTransport(conn)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapCerts(rpcConn.Bootstrap(ctx))

	cl := &CertsCapnpClient{
		connection: rpcConn,
		capability: capability,
	}
	return cl, nil
}
