package capnpclient

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/caphelp"
)

// VerifyAuthzCapnpClient capnp client capability to verify authorization
type VerifyAuthzCapnpClient struct {
	capability hubapi.CapVerifyAuthz // capnp client of the authorization service
}

// Release this capability. To be invoked after use has completed.
func (authz *VerifyAuthzCapnpClient) Release() {
	authz.capability.Release()
}

// Verify if the client is authorized to access thingID for the operation
//func (verifyAuthz *VerifyAuthzCapnpClient) Verify(
//	ctx context.Context, clientID string, thingID string, authType string) (authorized bool) {
//
//	method, release := verifyAuthz.capability.VerifyAuthorization(ctx,
//		func(params hubapi.CapVerifyAuthz_verifyAuthorization_Params) error {
//			err2 := params.SetClientID(clientID)
//			_ = params.SetThingID(thingID)
//			_ = params.SetAuthType(authType)
//			return err2
//		})
//	defer release()
//	resp, err := method.Struct()
//	if err == nil {
//		authorized = resp.Authorized()
//		return authorized
//	}
//	return false
//}

func (authz *VerifyAuthzCapnpClient) GetPermissions(
	ctx context.Context, clientID string, thingAddr string) (permissions []string, err error) {

	method, release := authz.capability.GetPermissions(ctx,
		func(params hubapi.CapVerifyAuthz_getPermissions_Params) error {
			err := params.SetClientID(clientID)
			_ = params.SetThingAddr(thingAddr)
			return err
		})
	defer release()

	resp, err := method.Struct()
	if err == nil {
		permsCapnp, _ := resp.Permissions()
		permissions = caphelp.UnmarshalStringList(permsCapnp)
	}
	return permissions, err
}

func NewVerifyAuthzCapnpClient(cap hubapi.CapVerifyAuthz) *VerifyAuthzCapnpClient {
	verifyAuthz := &VerifyAuthzCapnpClient{
		capability: cap,
	}
	return verifyAuthz
}
