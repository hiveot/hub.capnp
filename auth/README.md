# auth service

## Objective

Provide single sign-on authentication with authorization for use by Hub clients on the local network.

Hub access from outside the local network is out of scope for this service. External access is handled through inter-domain hub-to-hub communication which has its own domain authentication mechanism.  

## Status

The status of this plugin is alpha. It is functional but breaking changes are expected.

## Audience

This project is aimed at IoT developers that value the security and interoperability that WoST brings. WoST Things are
more secure than traditional IoT devices as they do not run a server, but instead connect to a Hub to publish their
information and receive actions.

## Summary

This Hub service supports local authentication and authorization for use by services, IoT devices and end-users on the local network. This module manages:
1. the Hub CA certificate for generating signed certificates
2. the Hub server certificate for use by Hub services (currently a single server certificate for all Hub services)
3. Manage client certificates for use by Hub services (currently a single client certificate for all Hub services)
4. Manage client certificates for use by IoT client authentication as used by the idprov provisioning service
5. Manage users for user authentication and authorization
6. Manage groups for role based user access to Things that are provided by IoT devices 
7. Manage JWT access and refresh tokens for user authentication


### Hub CA Certificate

The WoST Hub utilizes a self-signed CA certificate. The Hub CA is an organization certificate and not a domain certificate, as it is intended for use on the local network.  

A self-signed CA certificate is generated on first startup. The CA certificate is used to generate an intermediary certificate which is in turn used to generate  

This certificate must be distributed to clients so they can validate the server and client certificates. 




### Hub Service Certificates 

Hub Services use a CA signed server certificate to authenticate themselves to clients. Clients to hub services include other Hub services, IoT devices, mobile client and web applications. These clients verify the certificate against the CA certificate.


### Certificate Based authentication

Hub services and IoT devices authenticate using certificates that are signed by the Hub CA. In case of Hub services:
- A self signed Hub CA is manually generated on first startup. It can be replaced by the administrator with another CA certificate and private key pair.
- A new Hub server TLS certificate is generated on each restart. Hub services that offer an API use this certificate for incoming TLS connections.  
- A new Hub client TLS certificate is also generated on each restart. Hub services that connect to other services use this certificate to authenticate themselves as a Hub service. Currently, all services use the same client certificate with full authorization. This implies that Hub services are fully trusted. 



1. A login API to create access and refresh token pair when providing valid client credentials
2. A UI with web service to login or refresh their token pair. 
3. An API to refresh the access/refresh token pair
4. A library to verify the access token using the CA certificate public key
5. An API to verify the access token
6. An API to authorize access to one or more 'Things' 

A typical scenario for a user login follows the oauth2 'client credential' grant type. No redirect is used. SPA web apps will request the   

Instead of automatic URL redirect when authentication fails, the client redirects itself and repeats the original request once tokens are obtained. This avoids the need to register a return URL with the auth service, which is not always possible in offline usage.   
1. User starts a web client
2. Web client accesses a Hub service API with the last known bearer access token in the authentication header
3. Hub service verifies the access token and responds
   1. If the access token is invalid, a 401 UnAuthenticated response is returned including an authentication URL of the hub auth service.
4. The client redirects to the auth service URL
5. The auth service presents a login screen for the user to enter credentials
6. On success, the auth service returns an access token and stores a new refresh token in a secure cookie. This token is not accessible to the client.
7. The client retries the original hub service API with the new access token.

A commandline utility is provided for manual user management, changing passwords (by administrator) and generate certificates.

### Certificate Generation

The hubauth module generates the following certificates and stores then in the Hub 'certs' folder.

1. The CA certificate and private key is generated on first startup, as a self-signed certificate. The CA certficate is
   shared with every client and server and is used to verify authenticity of all other certificates. The CA key is only
   used to generate the certificates listed here. Intermediary certificates are not used by WoST Hub services.
2. The hub certificate and key are regenerated on each startup using the CA certificate. This certificate is used by hub
   services such as the provisioning server (idprov), directory server (thingdir), and mosquitto message bus server. The
   keys of this certificate is used to generate and validate access and refresh tokens.
3. The plugin certificate and key are regenerated on each startup using the CA certificate. This client certificate is
   used by hub services to authenticate access other hub services, like for example the message bus.
4. IoT Device client certificates are generated by the IDProv provisioning server using the client's public key. These
   certificates include the OU 'things' and are only issued to IoT devices that publish Thing information.

A commandline utility can generate custom certificates for special users such as administrators. All services accept
certificates signed by the CA.

### Client Certificate Authentication

Authentication through client certificates is handled by the TLS protocol itself. A valid certificate is required in
order to make a connection. The certificate CN contains the user or service name, the OU is used to indicate the role in
the organization.

* Hub plugins use the hub client certificate to access other Hub services. The Hub plugin certificate contains 'plugin'
  as the OU which is used to authorize full access.
* IoT devices have the device-ID as the CN and 'iotdevice' as the OU. IoT devices are publishers of things and have
  access to all Things of which they are the publisher.
* The administrator certificate has their user loginID as the CN and 'admin' as the OU.
* Other clients have their loginID as the CN and 'client' as the OU.

Properly signed certificates are always accepted as valid authentication. Their OU are used in authorization.

### Username/password Authentication

Unpw authentication is used for users that connect through a web client or mobile application. Clients use the auth API
with their login ID and password to obtain a pair of tokens that are needed to use the service API. To change a password
a valid password is also needed.

Administrators manage users and passwords through the 'auth' commandline utility or through the auth API. Creating users
or resetting passwords requires an 'admin' client certificate authentication.

#### **Password Storage**

Passwords are stored by the service using argon2id hashing. This is chosen as it is one of the strongest hash algorithms
that is resistant to GPU cracking attacks and side channel
attacks. [See wikipedia](https://en.wikipedia.org/wiki/Argon2). In future other algorithms can be supported as needed.

#### **Token Generation**

Unpw authentication uses JWT with H256 hashing for access and refresh token generation. Hashing protocol negotiation is
not
allowed [as this is considered a weakness](https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/)
. The required hashing algorithm can be changed without notice but is dictated by the server. JWT is chosen as it
secures a set of 'claims' which includes the userID. The userID is needed for authorization of a request. The hub
private key is used as the secret shared amongst plugin services. It is regenerated on each restart of the hub (
requiring a new login). The JWT header includes the hashing algorithm which MUST match the algorithm service is
configured for before token verification takes place.

Clients request an access/refresh token-pair from the auth service endpoint by providing a login ID and a secret (
password). The refresh token can be used to obtain a new set of access/refresh tokens. This approach allows for
stateless verification of the access token for its (configurable) validity period with a default of 1 hour. The refresh
token is valid for a longer with a default of two weeks. After the token expires, the user has to login to obtain a new
token pair. Until then no login is required. Refresh tokens are stored in secure client cookies which means they are not
accessible by clients and each applicatino must must login separately. Once logged in refresh tokens are used to avoid
unnecesary logins.

Access and refresh tokens include claim with the IP address of the sender, which must match during verification. Any
attempt to use the tokens with a different IP fails. The refresh token will be invalidated if an IP mismatch is detected
even if it hasn't expired yet.

All Hub services accept the access token within its validity period. Services verify the access token and its claims
using the server certificate key which all services have access to. Since a restart of the hub regenerates this key,
access and refresh tokens are invalidated at a Hub restart.

In short, services perform stateless verification of access tokens and return unauthorized if the token is expired or
doesn't match the IP address of the sender. Clients must refresh their tokens if they receive an unauthorized response
from one of the services and retry the request.

Weakness: Access to the tokens is the achilles heel of this approach. If a bad actor obtains a token while it is still
valid, and can spoof its IP address to that of the token, then security is compromised.

## Authorization

The auth module provides a library to assist services with handling authorizing requests.

Hub authorization groups things together and lets users view or control those things. This makes is easy to give
multiple users access to the same set of Things and at the same time simplifies access management of things.

Hub authorization is based on roles in groups that are centrally managed by the auth module:

1. Plugins have full authorization and access to all Things on the Hub
2. Users only have access to Things in the same group, based on their role:
    - view role gives read access to Things
    - control role lets the user control things
    - manage role lets the user configure things
    - admin role has full access
    - thing role is for things only and lets the thing manage itself

The 'all' group is built-in and automatically includes all Things. To allow a user to view all Things, the loginID is
added to the all group with the 'view' role.

A future considerations is to automatically create groups based on Thing Type, for example a group of environmental
sensors. This further simplifies group creation as things are automatically added to the groups they serve. This
requires a good consistent vocabulary of Thing types which is still tbd.

### Group Management

Things, users, groups and roles are defined in the ACL store. The first store implementation is a file that is loaded in
memory. The auth commandline lets the administrator add and remove users from the group. A REST API for managing groups
is planned.

The client library automatically reloads the file if it is modified.

To authorize a request, the client library needs to know the login-ID of the user, thing-ID to access, read/write
operation to perform and in case of writing the message type to write. These write message types are the Thing
Description document (TD), values update, action, event, and configuration.

The role permissions for these actions are:

| Role   |  TD   | Configure | Values  |  Event | Action
|--------|-------| --------- | ------- | -------| -------
| view   | read  | -         | read    | read   | -
| control| read  | -         | read    | read   | write
| manage | read  | write     | read    | read   | write
| admin  | read  | write     | read    | read   | write
| thing  | write | read      | write   | write  | write
| plugin | write | write     | write   | write  | write

## Build and Installation

### Build & Install (tentative)

Run 'make all' to build and 'make install' to install as a user.

See [hub's README.md](https://github.com/wostzone/hub/README.md) for more details.

## Usage

Code below is pseudocode and needs to be updated.

### User login

```golang
  authClient := NewAuthClient(address, port)
accessToken, err := authClient.Login(username, password)
// login sets the refresh token in a secure cookie. 
```

### User refresh auth tokens

```golang
  authClient := NewAuthClient(address, port)
// The refresh token was stored in a secure cookie. 
accessToken, err := authClient.Refresh()
if err != nil {
// if token cannot be refreshed then request user login    
}
```

### Server validates token (authenticate)

Access tokens can be validated stateless as long as we know to verification secret

```golang
  secret := ReadPrivateKey(file)
authClient := NewAuthClient(secret)
claims, err := authClient.VerifyToken(accessToken)
if err != nil {
return unauthorized
}
```

### Server authorizes user

Authorization keeps the ACL store in memory.

* loginID is the user login ID when authenticated with password
* ou is the user OU when authenticated using a client certificate
* thingID is the thing to access
* writing is read (false) or write (true) access
* writeType is the type of record to write, eg TD, Action, Event, Config, ...

```golang
  thing1ID := "urn:zone1:publisher1:thing1"
loginID := "user1"
ou := ""   // from client certificate if used
writing := false
writeType := MessageTypeTD
aclStore := aclstore.NewAclFileStore(aclFilePath, PluginID)
az := authorize.NewAuthorizer(aclStore)
az.VerifyAuthorization(loginID, ou, thingID, writing, writeType)
```

### Administrator adds user to group

To allow a user to view things in group 'temperature'

```golang
  groupName := "temperature"
aclStore := aclstore.NewAclFileStore(aclFilePath, PluginID)
aclStore.SetRole(loginID, groupName, authorize.GroupRoleViewer)
```

Or editing the acl file directly:

> groups.yaml

```yaml
all:
  admin: GroupRoleManager

temperature:
  user1: GroupRoleViewer
  urn:zone1:publisher1:thing1: ThingRole
  urn:zone1:publisher1:thing2: ThingRole
```
