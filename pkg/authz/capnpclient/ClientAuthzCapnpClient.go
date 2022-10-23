package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
)

// ClientAuthzCapnpClient capnp client capability to verify authorization
type ClientAuthzCapnpClient struct {
	capability hubapi.CapClientAuthz
}

func (authz *ClientAuthzCapnpClient) GetPermissions(
	ctx context.Context, thingID string) (permissions []string, err error) {

	method, release := authz.capability.GetPermissions(ctx,
		func(params hubapi.CapClientAuthz_getPermissions_Params) error {
			params.SetThingID(thingID)
			return nil
		})
	defer release()

	resp, err := method.Struct()
	if err == nil {
		permsCapnp, _ := resp.Permissions()
		permissions = caphelp.CapnpToStrings(permsCapnp)
	}
	return permissions, err
}

func NewClientAuthzCapnpClient(cap hubapi.CapClientAuthz) *ClientAuthzCapnpClient {
	clientAuthz := &ClientAuthzCapnpClient{
		capability: cap,
	}
	return clientAuthz
}
