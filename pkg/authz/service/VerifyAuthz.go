package service

import (
	"context"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/service/aclstore"
)

// VerifyAuthz determines client authorization for access to Things
// Authorization uses access control lists with group membership and roles to determine if a client
// is authorized to receive or post a message. This applies to all users of the message bus,
// regardless of how they are authenticated.
type VerifyAuthz struct {
	aclStore *aclstore.AclFileStore
}

// Release this client capability. To be invoked after use has completed.
func (vauthz *VerifyAuthz) Release() {
	// nothing to do here
}

// VerifyAuthorization tests if the client has access to the device for the given operation
// The thingID is implicitly included in the 'all' group. Members of the 'all' group can access all things
// based on their role in that group.
//
//  clientID is the deviceID, service-ID, or login name of the client seeking permission
//  thingID is the ID of the Thing to access
//  authType is one of: AuthEmitAction, AuthPubEvent, AuthPubTD, AuthReadTD, AuthWriteProperty
//
// This returns false if access is denied or true if authorized
// func (clauthz *VerifyAuthz) VerifyAuthorization(ctx context.Context,
// 	clientID string, thingID string, authType string) bool {

// 	logrus.Debugf("clientID='%s' thingID='%s', authType='%s'", clientID, thingID, authType)

// 	clientRole := clauthz.aclStore.GetRole(ctx, clientID, thingID)

// 	// pre-checks base on the cient's role

// 	// IoT devices are restricted to access things they are a publisher of.
// 	if clientRole == authz.ClientRoleIotDevice {
// 		if !clauthz.IsPublisher(ctx, clientID, thingID) {
// 			msg := fmt.Sprintf("Refused access by device '%s' to thingID '%s'. Thing belongs to a different publisher", clientID, thingID)
// 			logrus.Info(msg)
// 			return false
// 		}
// 	}
// 	allowed := clauthz.VerifyRolePermission(ctx, clientRole, authType)
// 	logrus.Debugf("role ('%s') permission check for (authType=%s): %t", clientRole, authType, allowed)
// 	return allowed
// }

// VerifyRolePermission determine if the consumer role allows the read/write operation.
// This is currently hard-coded to verify access to/by Things.
//
// When new services require additional roles or permissions in the future, this can be changed to
// accomodate those use-cases.
//
// permType describes permission to verify: permEmitAction, permPubEvent, ...
//
// Returns true if permission is denied, nil if granted
// func (clauthz *VerifyAuthz) VerifyRolePermission(ctx context.Context, role string, permType string) bool {
// 	_ = ctx
// 	switch permType {

// 	case authz.PermEmitAction:
// 		return role == authz.ClientRoleOperator ||
// 			role == authz.ClientRoleManager

// 	case authz.PermPubTD:
// 		return role == authz.ClientRoleIotDevice

// 	case authz.PermReadAction:
// 		return role == authz.ClientRoleOperator ||
// 			role == authz.ClientRoleManager

// 	case authz.PermReadEvent:
// 		return role == authz.ClientRoleOperator ||
// 			role == authz.ClientRoleManager ||
// 			role == authz.ClientRoleViewer

// 	case authz.PermReadTD:
// 		return role == authz.ClientRoleOperator ||
// 			role == authz.ClientRoleManager ||
// 			role == authz.ClientRoleViewer

// 	default:
// 		return false
// 	}
// }

// GetPermissions returns a list of permissions a client has for a Thing
func (vauthz *VerifyAuthz) GetPermissions(
	ctx context.Context, clientID string, thingID string) (permissions []string, err error) {

	clientRole := vauthz.aclStore.GetRole(ctx, clientID, thingID)
	switch clientRole {
	case authz.ClientRoleIotDevice:
		permissions = []string{authz.PermReadAction, authz.PermPubEvent, authz.PermPubTD}
	case authz.ClientRoleManager:
		permissions = []string{authz.PermEmitAction, authz.PermReadEvent, authz.PermReadAction,
			authz.PermReadTD, authz.PermWriteProperty}
	case authz.ClientRoleOperator:
		permissions = []string{authz.PermEmitAction, authz.PermReadEvent, authz.PermReadAction, authz.PermReadTD}
	case authz.ClientRoleViewer:
		permissions = []string{authz.PermReadEvent, authz.PermReadAction, authz.PermReadTD}
	default:
		permissions = []string{}
	}
	return permissions, nil
}

// IsPublisher checks if the deviceID is the publisher of the thingID.
// This requires that the thingID is formatted as "urn[:zone][:publisherID]:thingID...""
// Returns true if the deviceID is the publisher of the thingID, false if not.
func (vauthz *VerifyAuthz) IsPublisher(ctx context.Context, deviceID string, thingID string) (bool, error) {
	_ = ctx

	zone, publisherID, thingDeviceID, deviceType := thing.SplitThingID(thingID)
	_ = zone
	_ = thingDeviceID
	_ = deviceType
	return publisherID == deviceID, nil
}

// NewVerifyAuthz creates an instance handler for verifying authorization.
//  aclStore provides the functions to read and write authorization rules
func NewVerifyAuthz(aclStore *aclstore.AclFileStore) *VerifyAuthz {
	vauthz := VerifyAuthz{
		aclStore: aclStore,
	}
	return &vauthz
}
