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

// Release this capability. To be invoked after use has completed.
func (clAuthz *ClientAuthzCapnpClient) Release() {
	clAuthz.capability.Release()
}

func (clAuthz *ClientAuthzCapnpClient) GetPermissions(
	ctx context.Context, thingID string) (permissions []string, err error) {

	method, release := clAuthz.capability.GetPermissions(ctx,
		func(params hubapi.CapClientAuthz_getPermissions_Params) error {
			params.SetThingID(thingID)
			return nil
		})
	defer release()

	resp, err := method.Struct()
	if err == nil {
		permsCapnp, _ := resp.Permissions()
		permissions = caphelp.UnmarshalStringList(permsCapnp)
	}
	return permissions, err
}

func NewClientAuthzCapnpClient(cap hubapi.CapClientAuthz) *ClientAuthzCapnpClient {
	clientAuthz := &ClientAuthzCapnpClient{
		capability: cap,
	}
	return clientAuthz
}
