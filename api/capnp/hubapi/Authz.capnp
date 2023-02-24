# Cap'n proto definition for the authorization service
@0xae2da827da0eecef;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub/api/go/hubapi");

const authzServiceName :Text = "authz";
# Service name for use in logging and connectivity

#---------------------------------------------------
# Client roles set client permissions for accessing Things that are members of the same group
# The mapping of roles to operations is currently hard coded aimed at managing Things.
# For simplicity, clients roles are hierarchical and only a single role per group is supported.

const clientRoleNone :Text = "none";
# GroupRoleNone indicates that the client has no particular role. It can not do anything until
# the role is upgraded to viewer or better.
#  Read permissions: none
#  Write permissions: none

const clientRoleViewer :Text = "viewer";
# ClientRoleViewer lets a client subscribe to Thing TD and Thing Events
#  Read permissions: readTDs, readEvents
#  Write permissions: none

const clientRoleOperator :Text= "operator";
# ClientRoleOperator lets a client subscribe to Thing TD, events and publish actions
#  Read permissions: readTDs, readEvents, readActions
#  Write permissions: emitAction

const clientRoleManager :Text = "manager";
# ClientRoleManager lets a client subscribe to Thing TD, events, publish actions and update configuration
#  Read permissions: readTDs, readEvents, readActions
#  Write permissions: emitAction, writeProperty

const clientRoleIotDevice :Text = "iotdevice";
# ClientRoleIotDevice for IoT devices that read/write for things it is the publisher of.
# IoT Devices can publish events and updates for Things it the publisher of. This is determined
# by the publisherID part of the thingAddr.
#  Read permissions: readActions
#  Write permissions: pubTD, pubEvent, emitAction



#--- Permissions that can be authorized ---
# The list of permissions is currently hard coded aimed at managing TD's
# It is expected that future services will add permissions but that is for later.

const permEmitAction :Text = "permEmitAction";
# PermEmitAction permission to emit an action

const permPubEvent :Text = "permPubEvent";
# PermPubEvent permission to publish events, including property value events

const permPubTD :Text = "permPublishTD";
# PermPubTD permission to publish a TD document

const permReadAction :Text = "permReadAction";
# PermReadActions permission to read thing actions 

const permReadEvent :Text = "permReadEvent";
# PermReadEvents permission to read thing events 

const permReadTD :Text = "permReadTD";
# PermReadTD permission to read TD documents

const permWriteProperty :Text = "permWriteProperty";
# PermWriteProperty permission to write a thing property (configuration) value


const allGroupName :Text = "all";
# The all group implicitly contains all Things as a member

struct RoleMap {
    # map of key:role pairs
    entries @0 :List(Entry);
    struct Entry {
        key @0 :Text;
        role @1 :Text;
    }
}

struct Group {
# Group containing members and their roles

    name @0 :Text;
    # group name

    memberRoles @1 :RoleMap;
    # Map of member roles. Unfortunately capnp doesn't do maps, so return a list
}


const capNameClientAuthz :Text = "capClientAuthz";
const capNameManageAuthz :Text = "capManageAuthz";
const capNameVerifyAuthz :Text = "capVerifyAuthz";

interface CapAuthz {
# CapAuthz defines the interface of the authorization service

	capClientAuthz @0 (clientID :Text) -> (cap :CapClientAuthz);
	# Get the capability to verify a client's authorization

    capManageAuthz @1 (clientID :Text) -> (cap :CapManageAuthz);
    # Get the capability to manage authorization groups

	capVerifyAuthz @2 (clientID :Text) -> (cap :CapVerifyAuthz);
	# Get the capability to verify authorization
}


interface CapClientAuthz  {
# CapClientAuthz defines the capability to verifying authorization of a client.
# Intended for services or clients

    # GetPermissions returns a list of the permissions the client has for a Thing
    getPermissions @0 (thingAddr :Text) -> (permissions :List(Text));
}

interface CapManageAuthz {
# CapManageAuthz defines the capability for managing authorization groups.
# Intended for use by administrators.
	addThing @0 (thingAddr :Text, groupName :Text) -> ();
	# AddThing adds a Thing to a group

    getGroup @1 (groupName :Text) -> (group :Group);
    # GetGroup returns the group with the given name or an error if the group is not found

    getGroupRoles @2 (clientID :Text) -> (roles :RoleMap);
    # getGroupRoles returns the list of group-roles the client is a member of

    listGroups @3 (limit :Int32, offset :Int32) -> (groups :List(Group));
    # ListGroups returns the list of known groups

    removeAll @4 (clientID :Text) -> ();
    # RemoveAll removes a client or thing from all groups

	removeClient @5 (clientID :Text, groupName :Text) -> ();
	# RemoveClient removes a client from a group

	removeThing @6 (thingAddr :Text, groupName :Text) -> ();
	# RemoveThing removes a Thing from a group

	setClientRole @7 (clientID :Text, groupName :Text, role :Text) -> ();
	# SetClientRole sets the role for the client in a group.
	# Note that 'things' are also clients. Things are added to groups with the role ClientRoleThing
	#
	# If the client is not a member of a group the client will be added.
	# If the client is already a member of the group, its role will be replaced by the given role.
}


interface CapVerifyAuthz  {
# CapVerifyAuthz defines the capability for verifying authorization.
# Intended for services that provide access to Thing information.

	#verify @0 (clientID :Text, thingAddr :Text, authType :Text) -> (authorized :Bool);
	# VerifyAuthorization verifies the client's authorization to access a Thing
	#  clientID to authorize
	#  thingAddr of the resources to access
	#  authType of the permission to verify
	# Returns true if permission is granted, false if denied

    getPermissions @0 (clientID :Text, thingAddr :Text) -> (permissions :List(Text));
    # GetPermissions returns a list of the permissions the client has for a Thing
    #  clientID is the login ID of the user
}
