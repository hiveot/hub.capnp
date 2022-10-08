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
	ctx        context.Context
}

// CapManageProvisioning provides the capability to manage provisioning requests
func (cl *ProvisioningCapnpClient) CapManageProvisioning() provisioning.IManageProvisioning {
	getCap, _ := cl.capability.CapManageProvisioning(cl.ctx, nil)
	capability := getCap.Cap()
	return NewManageProvisioningCapnpClient(capability)
}

// CapRequestProvisioning provides the capability to provision IoT devices
func (cl *ProvisioningCapnpClient) CapRequestProvisioning() provisioning.IRequestProvisioning {
	getCap, _ := cl.capability.CapRequestProvisioning(cl.ctx, nil)
	capability := getCap.Cap()
	return NewRequestProvisioningCapnpClient(capability)
}

// CapRefreshProvisioning provides the capability for IoT devices to refresh
func (cl *ProvisioningCapnpClient) CapRefreshProvisioning() provisioning.IRefreshProvisioning {
	getCap, _ := cl.capability.CapRefreshProvisioning(cl.ctx, nil)
	capability := getCap.Cap()
	return NewRefreshProvisioningCapnpClient(capability)
}

// NewProvisioningCapnpClient returns a provisioning service client using the capnp protocol
//  ctx is the context for retrieving capabilities
//  conn is the connection with the provisioning capnp RPC server
func NewProvisioningCapnpClient(ctx context.Context, connection net.Conn) (*ProvisioningCapnpClient, error) {
	var cl *ProvisioningCapnpClient

	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapProvisioning(rpcConn.Bootstrap(ctx))

	cl = &ProvisioningCapnpClient{
		connection: rpcConn,
		capability: capability,
		ctx:        ctx,
	}
	return cl, nil
}
