package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/provisioning"
)

// ProvisioningCapnpClient provides a POGS wrapper around the generated provisioning capnp client
// This implements the IProvisioning interface
type ProvisioningCapnpClient struct {
	connection *rpc.Conn              // connection to the capnp server
	capability hubapi.CapProvisioning // capnp client
}

// CapManageProvisioning provides the capability to manage provisioning requests
func (cl *ProvisioningCapnpClient) CapManageProvisioning(ctx context.Context) provisioning.IManageProvisioning {
	getCap, release := cl.capability.CapManageProvisioning(ctx, nil)
	defer release()
	capability := getCap.Cap()
	return NewManageProvisioningCapnpClient(capability.AddRef())
}

// CapRequestProvisioning provides the capability to provision IoT devices
func (cl *ProvisioningCapnpClient) CapRequestProvisioning(ctx context.Context) provisioning.IRequestProvisioning {
	getCap, release := cl.capability.CapRequestProvisioning(ctx, nil)
	defer release()
	capability := getCap.Cap()
	return NewRequestProvisioningCapnpClient(capability.AddRef())
}

// CapRefreshProvisioning provides the capability for IoT devices to refresh
func (cl *ProvisioningCapnpClient) CapRefreshProvisioning(ctx context.Context) provisioning.IRefreshProvisioning {
	getCap, release := cl.capability.CapRefreshProvisioning(ctx, nil)
	defer release()
	capability := getCap.Cap()
	return NewRefreshProvisioningCapnpClient(capability.AddRef())
}

// Release the client capability
func (cl *ProvisioningCapnpClient) Release() {
	cl.capability.Release()
}

// NewProvisioningCapnpClient returns a provisioning service client using the capnp protocol
//
//	ctx is the context for this client's connection. Release it to release the client.
//	conn is the connection with the provisioning capnp RPC server
func NewProvisioningCapnpClient(ctx context.Context, connection net.Conn) (*ProvisioningCapnpClient, error) {
	var cl *ProvisioningCapnpClient

	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapProvisioning(rpcConn.Bootstrap(ctx))

	cl = &ProvisioningCapnpClient{
		connection: rpcConn,
		capability: capability,
	}
	return cl, nil
}
