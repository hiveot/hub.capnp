package capnpserver

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/caphelp"
	"github.com/hiveot/hub/pkg/authz"
)

// VerifyAuthzCapnpServer provides the capnp RPC server for Client authorization
type VerifyAuthzCapnpServer struct {
	srv authz.IVerifyAuthz
}

func (capsrv *VerifyAuthzCapnpServer) GetPermissions(
	ctx context.Context, call hubapi.CapVerifyAuthz_getPermissions) (err error) {

	args := call.Args()
	clientID, _ := args.ClientID()
	thingAddr, _ := args.ThingAddr()
	permissions, err := capsrv.srv.GetPermissions(ctx, clientID, thingAddr)
	if err == nil {
		res, err2 := call.AllocResults()
		err = err2
		_ = res.SetPermissions(caphelp.MarshalStringList(permissions))
	}
	return err
}
