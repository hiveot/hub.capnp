# authn service

## Objective

Provide authorization to access resources based on the client role and group. 

## Summary

This Hub service supports local user authorization for use by clients such as users, hub services and administrators. Users are identified by the user-ID, services and administrators by the CN in their certificate.

Authenticated clients have a role in each group they are a member of. Authorization to an IoT resource in the same group is granted based on the role of the client.

* clients can be users, services, and IoT devices. They must be authenticated using a valid certificate or access token.
* groups contain resources and clients. Clients can access resources in the same group based on their role. A client only has a single role.
  * The 'all' group includes all resources without need to add them explicitly. Use with care. 
* role. Clients have a role in a group. The role determines the action the client is allowed on the resource ('Things'). Roles are:
  * viewer: allows read-only access to the resource attributes such as Thing properties and output values
  * operator: in addition to viewer, allows operating the resource inputs such as a Thing switch
  * manager: in addition to operator, allows changing the resource configuration
  * administrator: in addition to manager, can manage users to the group
  * thing: role is for use by IoT devices only and identifies it as the resource to access. Thing publishers are devices that have full access to the Things they publish. They are identified by their publisher ID in the device client certificate. 
  
The 'all' group is built-in and automatically includes all Things. To allow a user to view all Things, the loginID is added to the all group with the 'view' role.

A future considerations is to automatically add things to groups based on their Thing Type, for example a group of environmental sensors. This further simplifies group creation as things are automatically added to the groups they serve. This requires a good consistent vocabulary of Thing types which is still tbd.

### Group Management

Things, users, groups and roles are defined in the ACL (access control list) store. The default store implementation is file based that is loaded in memory. The 'authz' commandline lets the administrator manage users, groups and roles in this file. A REST API is planned.

The client library automatically reloads the file if it is modified.

To authorize a request, the authz library uses the ID of the client to determine the role for the requested resource(s). The role determines the allowed Thing actions. Thing actions are:
* 'Read TD'. On request return the full TD of the Thing.
* 'Configure'. Publish a configuration message for the Thing. If accepted by the Thing publisher then this results in an updated TD being published. The Thing TD is never modified directly. Only managers and operators are allowed to publish a Thing Configuration message. 
* 'Event' is published by the Thing publisher when an event happens. Only Thing publishers are allowed to publish event messages.  All members of a group are allowed to subscribe to events from Things in that group.
* 'Action' is a message published by client to operate a thing. For example control a switch. Viewers are not allowed to publish this message.

The role permissions for these message actions are:

| Role / action | Read TD | Configure | Event  | Action |
|---------------|---------|-----------| ------ |--------|
| viewer        | read    | -         | read   | -      |
| operator      | read    | -         | read   | write  |
| manager       | read    | write     | read   | write  |
| admin         | read    | write     | read   | write  |
| thing         | write   | write     | write  | write  |

## Build and Installation

### Build & Install (tentative)

Run 'make all' to build and 'make install' to install as a user.

See [hub's README.md](https://github.com/wostzone/hub/README.md) for more details.

## Usage

Examples below are tentative pseudocode.

### Administrator adds user to group

To allow a user to view things in group 'temperature'

From the CLI:
```bash
authz setrole {userID} temperature viewer
```

In code:
```go
groupID := "temperature"
aclStore := aclstore.NewAclFileStore(aclFilePath, PluginID)
aclStore.SetRole(userID, groupID, authorize.GroupRoleViewer)
```

Or editing the acl file directly:

> groups.yaml

```yaml
all:
  admin: manager

temperature:
  user1: viewer
  urn:zone1:publisher1:thing1: thing
  urn:zone1:publisher1:thing2: thing
```



### Verify if a user has write access to a resource

In code:
```go
thingID := "urn:zone1:publisher1:thing1"
clientID := "user1"
ou := ""   // from client certificate if used
writing := false
writeType := MessageTypeTD
aclStore := aclstore.NewAclFileStore(aclFilePath, PluginID)

az := authz.NewAuthorizer(aclStore)
az.VerifyAuthorization(clientID, ou, thingID, writing, writeType)
```
