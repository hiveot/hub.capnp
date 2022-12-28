package authz

import (
	"context"
)

// ServiceName of the service used for logging and connection
const ServiceName = "authz"

// Client roles set permissions for operations on Things that are members of the same group
// The mapping of roles to operations is currently hard coded aimed at managing Things
const (
	// ClientRoleNone indicates that the client has no particular role. It can not do anything until
	// the role is upgraded to viewer or better.
	//  Read permissions: none
	//  Write permissions: none
	ClientRoleNone = "none"

	// ClientRoleIotDevice for IoT devices that read/write for things it is the publisher of.
	// IoT Devices can publish events and updates for Things it the publisher of. This is determined
	// by the deviceID that is included in the thingID.
	//  Read permissions: readActions
	//  Write permissions: pubTD, pubEvent, emitAction
	ClientRoleIotDevice = "iotdevice"

	// ClientRoleManager lets a client subscribe to Thing TD, events, publish actions and update configuration
	//  Read permissions: readTDs, readEvents, readActions
	//  Write permissions: emitAction, writeProperty
	ClientRoleManager = "manager"

	// ClientRoleOperator lets a client subscribe to Thing TD, events and publish actions
	//  Read permissions: readTDs, readEvents, readActions
	//  Write permissions: emitAction
	ClientRoleOperator = "operator"

	// ClientRoleThing identifies the client as a Thing
	// Things can publish events and updates for themselves.
	//  Read permissions: readActions
	//  Write permissions: pubTD, pubEvent, emitAction
	ClientRoleThing = "thing"

	// ClientRoleViewer lets a client subscribe to Thing TD and Thing Events
	//  Read permissions: readTDs, readEvents
	//  Write permissions: none
	ClientRoleViewer = "viewer"
)

// Permissions that can be authorized
// The list of permissions is currently hard coded aimed at managing Things
// It is expected that future services will add permissions but that is for later.
const (
	// PermEmitAction permission of emitting an action
	PermEmitAction = "permEmitAction"

	// PermPubEvent permission to publish events, including property value events
	PermPubEvent = "permPubEvent"

	// PermPubTD permission to publish a TD document
	PermPubTD = "permPubTD"

	// PermReadAction permission of read/subscribe to actions
	PermReadAction = "permReadAction"

	// PermReadEvents permission to read/subscribe to events
	PermReadEvent = "permReadEvent"

	// PermReadTD permission to read TD documents
	PermReadTD = "permReadTD"

	// PermWriteProperty permission to write a property (configuration) value
	PermWriteProperty = "permWriteProperty"
)

// AllGroupName is the built-in group containing all resources
const AllGroupName = "all"

// RoleMap for members or memberships
type RoleMap map[string]string // clientID:role, groupName:role

// Group is a map of clientID:role
type Group struct {
	Name string
	// map of clients and their role in this group
	MemberRoles RoleMap
}

// NewGroup creates an instance of a group with member roles
func NewGroup(groupName string) Group {
	return Group{
		Name:        groupName,
		MemberRoles: make(RoleMap),
	}
}

// IAuthz defines the interface of the authorization service
type IAuthz interface {

	// CapClientAuthz provides the capability to verify a client's authorization
	CapClientAuthz(ctx context.Context, clientID string) IClientAuthz

	// CapManageAuthz provides the capability to manage authorization groups
	CapManageAuthz(ctx context.Context, clientID string) IManageAuthz

	// CapVerifyAuthz provides the capability to verify authorization
	CapVerifyAuthz(ctx context.Context, clientID string) IVerifyAuthz
}

// IClientAuthz defines the capability for verifying authorization of a client.
type IClientAuthz interface {
	// Release this client capability after its use.
	Release()

	// GetPermissions returns the permissions the client has for a Thing
	// Returns an array of permissions, eg PermEmitAction, etc
	GetPermissions(ctx context.Context, thingAddr string) (permissions []string, err error)
}

// IManageAuthz defines the capability for managing authorization groups.
// Intended for use by administrators.
type IManageAuthz interface {
	// AddThing adds a Thing to a group
	AddThing(ctx context.Context, thingAddr string, groupName string) error

	// Release this client capability after its use.
	Release()

	// GetGroup returns the group with the given name, or an error if group is not found.
	// GroupName must not be empty and must be an existing group
	// Returns an error if the group does not exist.
	GetGroup(ctx context.Context, groupName string) (group Group, err error)

	// GetGroupRoles returns a map of group:role for groups the client is a member of.
	GetGroupRoles(ctx context.Context, clientID string) (roles RoleMap, err error)

	// ListGroups returns the list of known groups
	ListGroups(ctx context.Context, limit int, offset int) (groups []Group, err error)

	// RemoveAll removes a client or thing from all groups
	RemoveAll(ctx context.Context, clientID string) error

	// RemoveClient removes a client from a group
	RemoveClient(ctx context.Context, clientID string, groupName string) error

	// RemoveThing removes a Thing from a group
	RemoveThing(ctx context.Context, thingAddr string, groupName string) error

	// SetClientRole sets the role for the client in a group.
	// Note that 'things' are also clients. Things are added to groups with the role ClientRoleThing
	// If the client is not a member of a group the client will be added.
	// If the client is already a member of the group, its role will be replaced by the given role.
	SetClientRole(ctx context.Context, clientID string, groupName string, role string) error
}

// IVerifyAuthz defines the capability for verifying authorization.
// Intended for services that provide access to Thing information.
type IVerifyAuthz interface {
	// Release this client capability after its use.
	Release()

	// GetPermissions returns the permissions a client has for a Thing
	// Returns an array of permissions, eg PermEmitAction, etc
	GetPermissions(ctx context.Context, clientID, thingAddr string) (permissions []string, err error)
}
