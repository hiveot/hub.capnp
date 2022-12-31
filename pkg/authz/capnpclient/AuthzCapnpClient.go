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
func (authz *AuthzCapnpClient) CapClientAuthz(ctx context.Context, clientID string) (authz.IClientAuthz, error) {
	getCap, release := authz.capability.CapClientAuthz(ctx,
		func(params hubapi.CapAuthz_capClientAuthz_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	defer release()
	capability := getCap.Cap().AddRef()
	cl := NewClientAuthzCapnpClient(capability)
	return cl, nil
}

// CapManageAuthz provides the capability to manage authorization groups
func (authz *AuthzCapnpClient) CapManageAuthz(ctx context.Context, clientID string) (authz.IManageAuthz, error) {
	getCap, release := authz.capability.CapManageAuthz(ctx,
		func(params hubapi.CapAuthz_capManageAuthz_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	defer release()
	capability := getCap.Cap().AddRef()
	cl := NewManageAuthzCapnpClient(capability)
	return cl, nil
}

// CapVerifyAuthz provides the capability to verify authorization
// The type of client, OU of certificate, certsclient.OUService, OUIoTDevice, OUUser, OUAdmin
func (authz *AuthzCapnpClient) CapVerifyAuthz(ctx context.Context, clientID string) (authz.IVerifyAuthz, error) {

	getCap, release := authz.capability.CapVerifyAuthz(ctx,
		func(params hubapi.CapAuthz_capVerifyAuthz_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	defer release()
	capability := getCap.Cap().AddRef()
	cl := NewVerifyAuthzCapnpClient(capability)
	return cl, nil
}

func (authz *AuthzCapnpClient) Release() {
	authz.capability.Release()
}

// NewAuthzCapnpClient returns a authorization client using the capnp protocol
//
//	ctx is the context for retrieving capabilities
//	connection is the client connection to the capnp server
func NewAuthzCapnpClient(ctx context.Context, connection net.Conn) *AuthzCapnpClient {
	var cl *AuthzCapnpClient
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapAuthz(rpcConn.Bootstrap(ctx))

	cl = &AuthzCapnpClient{
		connection: rpcConn,
		capability: capability,
	}
	return cl
}
