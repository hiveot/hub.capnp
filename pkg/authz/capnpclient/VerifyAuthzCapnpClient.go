package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
)

// VerifyAuthzCapnpClient capnp client capability to verify authorization
type VerifyAuthzCapnpClient struct {
	capability hubapi.CapVerifyAuthz // capnp client of the authorization service
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
	ctx context.Context, clientID string, thingID string) (permissions []string, err error) {

	method, release := authz.capability.GetPermissions(ctx,
		func(params hubapi.CapVerifyAuthz_getPermissions_Params) error {
			params.SetClientID(clientID)
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

func (authz *VerifyAuthzCapnpClient) IsPublisher(
	ctx context.Context, deviceID string, thingID string) (isPub bool, err error) {
	isPub = false
	method, release := authz.capability.IsPublisher(ctx,
		func(params hubapi.CapVerifyAuthz_isPublisher_Params) error {
			params.SetDeviceID(deviceID)
			params.SetThingID(thingID)
			return nil
		})
	defer release()

	resp, err := method.Struct()
	if err == nil {
		isPub = resp.Ispub()
	}
	return isPub, err
}

func NewVerifyAuthzCapnpClient(cap hubapi.CapVerifyAuthz) *VerifyAuthzCapnpClient {
	verifyAuthz := &VerifyAuthzCapnpClient{
		capability: cap,
	}
	return verifyAuthz
}
