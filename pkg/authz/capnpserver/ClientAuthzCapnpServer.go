package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/authz"
)

// ClientAuthzCapnpServer provides the capnp RPC server for Client authorization
type ClientAuthzCapnpServer struct {
	srv authz.IClientAuthz
}

func (capsrv *ClientAuthzCapnpServer) GetPermissions(
	ctx context.Context, call hubapi.CapClientAuthz_getPermissions) (err error) {

	args := call.Args()
	thingID, _ := args.ThingID()
	permissions, err := capsrv.srv.GetPermissions(ctx, thingID)
	if err == nil {
		res, err2 := call.AllocResults()
		err = err2
		_ = res.SetPermissions(caphelp.StringsToCapnp(permissions))
	}
	return err
}
