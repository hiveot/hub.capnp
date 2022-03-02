# authn service

## Objective

Provide user based authentication for use by Hub clients on the local network where Internet access might or might not be available. Hub user access from outside the local network is not allowed to reduce security risks. The authn service has the option to refuse authentication requests from a client with an invalid source IP. 

### Scope

In-scope is to provide identity management for users on the local network. Logging into the authn service will provide credentials required to authorize access to Thing resources. 

Out of scope are:
* User authorization to access Things. Authentication is not sufficient to access Thing data without authorization.
* Authenticate from outside the local network. To share information with Hub's over the internet the architecture defines a hub bridge service. This service accepts incoming connections from local Hub's and shares Thing updates between connected Hubs. 
* A cloud based Hub that is hardened for user authentication over the internet. A cloud Hub is intended to aggregate information from other Hubs via the Hub bridge and allow users to access this information. Although a regular Hub can be used as cloud Hub, the step to harden its authentication sufficiently is currently not in scope for this auth service.
* IoT device authentication. IoT devices use a client certificate to authenticate via TLS. This is considered sufficiently secure to allow access from anywhere. Each IoT device has a dedicated certificate that is only authorized to access to the Thing information published by the device. A valid client certificate is required. 
* Hub service authentication. Hub services use client certificates to authenticate via TLS. A valid client certificate is required. As Hub services have full access to the Hub it is recommended to keep Hub services running locally on the Hub or at least on the local network.  

## Status

The status of this plugin is alpha. It is functional but breaking changes are expected.

## Audience

This project is aimed at IoT developers that value the security and interoperability that WoST brings. WoST Things are more secure than traditional IoT devices as they do not run a server, but instead connect to a Hub to publish their information and receive actions.

## Summary

This Hub service supports local user authentication for use by services, IoT devices and end-users on the local network. This module manages users and their credentials and issues JWT tokens in order to access resources.

Login ID/password authentication is used for users that connect through a web client or mobile application. Clients use the authn API with their login ID and password to obtain a pair of tokens that are needed to use the services. 

Administrators manage users and passwords through the 'authn' commandline utility or through the authn API. Creating users or resetting passwords requires an 'admin' client certificate.

#### **Password Storage**

Passwords are stored by the service using argon2id hashing. This is chosen as it is one of the strongest hash algorithms that is resistant to GPU cracking attacks and side channel attacks. [See wikipedia](https://en.wikipedia.org/wiki/Argon2). In future other algorithms can be supported as needed.

#### **Token Generation**

Authn uses JWT with H256 hashing for access and refresh token generation. Hashing protocol negotiation is not allowed [as this is considered a weakness](https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/). The hashing algorithm can be changed in future. 

An asymmetric public/private key pair is used to generate the access and refresh tokens. A public/private key pair is generated on startup. Services that verify access tokens must be given to the public key. Services on the Hub will look in the default location while services that run on a separate system should be provided the public key on installation. The service configuration enables the option to auto generate a new key pair on startup or to use the provided keys.

JWT is chosen as it includes a set of verified 'claims' which is needed to verify the userID. 

Clients request an access/refresh token-pair from the authn service endpoint by providing a login ID and a secret (password). The refresh token can be used to obtain a new set of access/refresh tokens. This approach allows for stateless verification of the access token for its (configurable) validity period with a default of 1 hour. The refresh token is valid for a longer time with a default of two weeks. When a refresh token has expired, the user has to login with credentials to obtain a new token pair. As long as the refresh token is renewed no login is required. Refresh tokens are stored in secure client cookies which means they are not accessible by javascript clients. 

Access and refresh tokens include claim with the IP address of the sender, which must match during verification. Any attempt to use the tokens with a different IP fails. The refresh token will be invalidated if an IP mismatch is detected even if it hasn't expired yet.

Hub services accept the access token within its validity period as long as they verify against the CA public key. The public key is available in the CA certificate or can be distributed as PEM file.

In short, services that accept access tokens perform stateless verification of the token against the CA public key and sender IP address. Clients must refresh their tokens if they receive an unauthorized response from one of the services and retry the request.

Weakness 1: Access to the tokens is the achilles heel of this approach. If a bad actor obtains an access token while it is still valid, and can spoof its IP address to that of the token, then security is compromised. This is somewhat mitigated by using TLS and requiring a valid server certificate, signed by the CA.


## Build and Installation

### Build & Install (tentative)

Run 'make all' to build and 'make install' to install as a user.

See [hub's README.md](https://github.com/wostzone/hub/README.md) for more details.

## Usage

Code below is pseudocode and needs to be updated.

### Add user and set password

Using the authn CLI. This utility should only be accessible to admin users:
```bash
 bin/authn adduser {userID}      # this will prompt for a password
 
 bin/authn deleteuser {userID}

 bin/authn setpasswd             # this will prompt for a password
```

Using the service API:
```golang
  // an administrator client certificate is required for this operation
  authnAdmin := NewAuthnAdmin(address, port, clientCert)
  
  err := authnAdmin.addUser(userID, passwd)

  err := authnAdmin.setPasswd(userID, passwd)
```

### User login

```golang
  authnClient := NewAuthnClient(address, port)
  accessToken, err := authnClient.Login(username, password)
  // login sets the refresh token in a secure cookie. 
```

### User refresh auth tokens

```golang
  authnClient := NewAuthnClient(address, port)
  // The refresh token was stored in a secure cookie. 
  accessToken, err := authnClient.Refresh()
  if err != nil {
    // if token cannot be refreshed then request user login    
  }   
```

### Server validates token (authenticate)

Access tokens can be validated using the verification public key 

```golang
  pubkey := ReadPublicKey(pubKeyFile)
  authenticator := NewJwtAuthenticator(pubkey)
  claims, err := authenticator.VerifyToken(accessToken)
  if err != nil {
    return unauthorized
  }
```

