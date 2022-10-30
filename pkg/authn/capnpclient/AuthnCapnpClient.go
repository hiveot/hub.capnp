package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
)

// AuthnCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IAuthn interface
type AuthnCapnpClient struct {
	connection *rpc.Conn       // connection to capnp server
	capability hubapi.CapAuthn // capnp client of the authentication service
}

// CapUserAuthn provides the authentication capability for authenticating users
func (cl *AuthnCapnpClient) CapUserAuthn(ctx context.Context, clientID string) authn.IUserAuthn {
	getCap, release := cl.capability.CapUserAuthn(ctx,
		func(params hubapi.CapAuthn_capUserAuthn_Params) error {
			params.SetClientID(clientID)
			return nil
		})
	defer release()
	capability := getCap.Cap().AddRef()
	return NewUserAuthnCapnpClient(capability)
}

func (cl *AuthnCapnpClient) CapManageAuthn(ctx context.Context) authn.IManageAuthn {
	getCap, release := cl.capability.CapManageAuthn(ctx, nil)
	defer release()
	capability := getCap.Cap().AddRef()
	return NewManageAuthnCapnpClient(capability)
}

// NewAuthnCapnpClient returns a authentication client using the capnp protocol
//  ctx is the context for retrieving capabilities
//  connection is the client connection to the capnp server
func NewAuthnCapnpClient(ctx context.Context, connection net.Conn) (*AuthnCapnpClient, error) {
	var cl *AuthnCapnpClient
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapAuthn(rpcConn.Bootstrap(ctx))

	cl = &AuthnCapnpClient{
		connection: rpcConn,
		capability: capability,
	}
	return cl, nil
}
