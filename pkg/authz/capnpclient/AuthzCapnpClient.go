// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authz"
)

// AuthzCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IAuthz interface
type AuthzCapnpClient struct {
	connection *rpc.Conn       // connection to capnp server
	capability hubapi.CapAuthz // capnp client of the authorization service
}

// CapClientAuthz provides the capability to verify a specific client's authorization
func (authz *AuthzCapnpClient) CapClientAuthz(ctx context.Context, clientID string) authz.IClientAuthz {
	getCap, _ := authz.capability.CapClientAuthz(ctx,
		func(params hubapi.CapAuthz_capClientAuthz_Params) error {
			params.SetClientID(clientID)
			return nil
		})
	capability := getCap.Cap()
	return NewClientAuthzCapnpClient(capability)
}

// CapManageAuthz provides the capability to manage authorization groups
func (authz *AuthzCapnpClient) CapManageAuthz(ctx context.Context) authz.IManageAuthz {
	getCap, _ := authz.capability.CapManageAuthz(ctx, nil)
	capability := getCap.Cap()
	return NewManageAuthzCapnpClient(capability)
}

// CapVerifyAuthz provides the capability to verify authorization
// The type of client, OU of certificate, certsclient.OUService, OUIoTDevice, OUUser, OUAdmin
func (authz *AuthzCapnpClient) CapVerifyAuthz(ctx context.Context) authz.IVerifyAuthz {
	getCap, _ := authz.capability.CapVerifyAuthz(ctx, nil)
	capability := getCap.Cap()
	return NewVerifyAuthzCapnpClient(capability)
}

// NewAuthzCapnpClient returns a authorization client using the capnp protocol
//  ctx is the context for retrieving capabilities
//  connection is the client connection to the capnp server
func NewAuthzCapnpClient(ctx context.Context, connection net.Conn) (*AuthzCapnpClient, error) {
	var cl *AuthzCapnpClient
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapAuthz(rpcConn.Bootstrap(ctx))

	cl = &AuthzCapnpClient{
		connection: rpcConn,
		capability: capability,
	}
	return cl, nil
}
