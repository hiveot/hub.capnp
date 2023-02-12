package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
)

// AuthnCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IAuthnService interface
type AuthnCapnpClient struct {
	capability hubapi.CapAuthn // capnp client of the authentication service
}

// CapUserAuthn provides the authentication capability for authenticating users
func (cl *AuthnCapnpClient) CapUserAuthn(ctx context.Context, clientID string) (authn.IUserAuthn, error) {
	getCap, release := cl.capability.CapUserAuthn(ctx,
		func(params hubapi.CapAuthn_capUserAuthn_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	defer release()
	capability := getCap.Cap().AddRef()
	return NewUserAuthnCapnpClient(capability), nil
}

func (cl *AuthnCapnpClient) CapManageAuthn(ctx context.Context, clientID string) (authn.IManageAuthn, error) {
	getCap, release := cl.capability.CapManageAuthn(ctx,
		func(params hubapi.CapAuthn_capManageAuthn_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	defer release()
	capability := getCap.Cap().AddRef()
	return NewManageAuthnCapnpClient(capability), nil
}

// Release this client capability
func (cl *AuthnCapnpClient) Release() {
	cl.capability.Release()
}

// NewAuthnClientFromCapnpConnection returns a new authentication client from a connection to a
// capnp server.
//
//	ctx is the context for retrieving capabilities
//	connection is the client connection to the capnp server
func NewAuthnClientFromCapnpConnection(ctx context.Context, connection net.Conn) *AuthnCapnpClient {
	var cl *AuthnCapnpClient
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapAuthn(rpcConn.Bootstrap(ctx))

	cl = &AuthnCapnpClient{
		capability: capability,
	}
	return cl
}

// NewAuthnClientFromCapnpCapability returns a authn client from its capnpCapability
// Use when using a proxy client such as the resolver and gateway.
func NewAuthnClientFromCapnpCapability(capability hubapi.CapAuthn) *AuthnCapnpClient {
	cl := &AuthnCapnpClient{
		capability: capability,
	}
	return cl
}
