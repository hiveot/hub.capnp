# authn service

## Objective

Provide user authentication on the local network. 

## Status

This service is functional but breaking changes can still be expected.


## Scope

In-scope is to provide identity management for users on the local network. Login to the authn service will provide tokens required to authorize access to Thing resources. 

Out of scope are:
* OAuth2 authentication. Since internet access is not guaranteed for this service, authentication with oauth2 might not work. Integration with auth0 or other oauth2 provider might be added in the future if a strong use-case arises. 
* Authorization is handled separately by the authz service.
* Authentication for a cloud based Hub. A cloud based Hub requires an authentication implementation that is hardened for access from the Internet. This authn service has not gone through the hardening process and it is therefore not recommended for this usage.
* IoT device authentication. IoT devices use a client certificate to authenticate via TLS. This is considered sufficiently secure to allow access from anywhere. Each IoT device has a dedicated certificate that is only authorized to access to the Thing information published by the device. A valid client certificate as provided by the idprov service is required. 
* Hub plugin/service authentication. Hub services use Unix Sockets locally and TLS connections with client certificates to connect over the network. A valid client certificate implies the user is authenticated.   

## Summary

This Hub service provides local user authentication on the local network. It manages users and their credentials and issues JWT tokens for use in HTTP authorization header.

Login ID/password authentication is used for users that connect through a web client or mobile application. Clients use the authn API with their login ID and password to obtain a pair of tokens that are required to use the Hub services. 

Administrators manage users and passwords through the 'hubcli' commandline utility or through the authn management API. Creating users or resetting passwords requires an 'admin' client certificate.

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


## Usage

Code below is pseudocode and needs to be updated.

### Add user and set password

Using the authn CLI. This utility should only be accessible to admin users:
```bash
 bin/hubapi authn adduser {userID}      # this will prompt for a password
 
 bin/hubapi authn deleteuser {userID}

 bin/hubapi authn setpasswd             # this will prompt for a password
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
