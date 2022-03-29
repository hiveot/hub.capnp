// Package authorize with definitions for group based authorization.
package authorize

// Group roles set permissions for accessing Things that are members of the same group
const (
	// GroupRoleNone indicates that the client has no particular role. It can not do anything until
	// the role is upgraded to viewer or better.
	// Subscribe permissions: none
	// Publish permissions: none
	GroupRoleNone = "none"

	// GroupRoleViewer lets a client subscribe to Thing TD and Thing Events
	// Subscribe permissions: TD, Events
	// Publish permissions: none
	GroupRoleViewer = "viewer"

	// GroupRoleOperator lets a client subscribe to Thing TD, events and publish actions
	// Subscribe permissions: TD, Events
	// Publish permissions: Actions
	GroupRoleOperator = "operator"

	// GroupRoleManager lets a client subscribe to Thing TD, events, publish actions and update configuration
	// Subscribe permissions: TD, Events
	// Publish permissions: Actions, Configuration
	GroupRoleManager = "manager"

	// GroupRoleThing indicates the client is a IoT device that can publish and subscribe
	// to Thing topics.
	// Things should only publish events and updates for Things it published the TD for.
	// Publish permissions: TD, Events
	// Subscribe permissions: Actions, Configuration
	GroupRoleThing = "thing"
)

// AuthGroup defines a group with Thing and Users
// The permission is determined by taking the thing permission and user permission and
// return the lowest of the two.
// Eg an admin role can do anything only if the thing allows it
//
// This allows for Things to be shared with other groups with viewing rights only, even though
// there are user or admins in that group.
// type AuthGroup struct {
// 	// The name of the group
// 	GroupName string
// 	// The members (thingIDs and userIDs) and their role: [memberid]role
// 	MemberRoles map[string]string
// }

// AclGroupAll defines the 'all' resources group
const AclGroupAll = "all"

// Authorization of actions
const (
	// AuthRead to read a TD
	AuthRead = "read"

	// AuthEmitAction to authorize emit an action
	AuthEmitAction = "emitAction"

	// AuthPubEvent to authorize publishing of events
	AuthPubEvent = "pubEvent"

	// AuthPubPropValue to authorize publishing of property values
	AuthPubPropValue = "pubPropValues"

	// AuthPubTD to authorize publishing of a TD document
	AuthPubTD = "pubTD"

	// AuthWriteProperty to authorize writing property value
	AuthWriteProperty = "writeProperty"
)
