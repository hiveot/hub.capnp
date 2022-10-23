package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
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
	thingID, _ := args.ThingID()
	permissions, err := capsrv.srv.GetPermissions(ctx, clientID, thingID)
	if err == nil {
		res, err2 := call.AllocResults()
		err = err2
		_ = res.SetPermissions(caphelp.MarshalStringList(permissions))
	}
	return err
}

func (capsrv *VerifyAuthzCapnpServer) IsPublisher(
	ctx context.Context, call hubapi.CapVerifyAuthz_isPublisher) (err error) {

	args := call.Args()
	deviceID, _ := args.DeviceID()
	thingID, _ := args.ThingID()
	isPub, err := capsrv.srv.IsPublisher(ctx, deviceID, thingID)
	if err == nil {
		res, err2 := call.AllocResults()
		err = err2
		res.SetIspub(isPub)
	}
	return err
}
