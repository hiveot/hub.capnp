// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"
	"net"
	"time"

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
	ctxCancel  context.CancelFunc
}

// CapDeviceCerts returns the capability to create device certificates
func (cl *CertsCapnpClient) CapDeviceCerts() certs.IDeviceCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, _ := cl.capability.CapDeviceCerts(cl.ctx, nil)
	capability := getCap.Cap()
	return NewDeviceCertsCapnpClient(capability)
}

// CapServiceCerts returns the capability to create service certificates
func (cl *CertsCapnpClient) CapServiceCerts() certs.IServiceCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, _ := cl.capability.CapServiceCerts(cl.ctx, nil)
	capability := getCap.Cap()
	return NewServiceCertsCapnpClient(capability)
}

// CapUserCerts returns the capability to create user certificates
func (cl *CertsCapnpClient) CapUserCerts() certs.IUserCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, _ := cl.capability.CapUserCerts(cl.ctx, nil)
	capability := getCap.Cap()
	return NewUserCertsCapnpClient(capability)
}

// CapVerifyCerts returns the capability to verify certificates
func (cl *CertsCapnpClient) CapVerifyCerts() certs.IVerifyCerts {

	// Get the capability for creating a device certificate for the given device
	getCap, _ := cl.capability.CapVerifyCerts(cl.ctx, nil)
	capability := getCap.Cap()
	return NewVerifyCertsCapnpClient(capability)
}

// Release the provided capabilities after use
// FIXME: Is this required, optional, or not useful at all?
// What is managing this lifecycle aspect?
func (cl *CertsCapnpClient) Release() {
	cl.capability.Release()
}

// NewCertServiceCapnpClient returns a capability to create certificates using the capnp protocol
// Intended for bootstrapping the capability chain
func NewCertServiceCapnpClient(address string, isUDS bool) (*CertsCapnpClient, error) {
	network := "tcp"
	if isUDS {
		network = "unix"
	}
	connection, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*60)
	capability := hubapi.CapCerts(rpcConn.Bootstrap(ctx))

	cl := &CertsCapnpClient{
		connection: rpcConn,
		capability: capability,
		ctx:        ctx,
		ctxCancel:  ctxCancel,
	}
	return cl, nil
}
