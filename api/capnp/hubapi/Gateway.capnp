# Cap'n proto definition for Hub gateway service
@0xdd3a962266ddd0e3;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub/api/go/hubapi");

using Resolver = import "Resolver.capnp";

const gatewayServiceName :Text = "gateway";


struct ClientInfo {
# ClientInfo contains client info as seen by the gateway
# Intended for diagnostics and troubleshooting

  clientID @0 :Text;
  # ClientID that is connected. loginID, serviceID, or IoT device ID

  authType @1 :Text;
  # AuthType identifies how the client is authenticated. See also the resolver
  #  AuthTypeUnauthenticated   - client is not authenticated
  #  AuthTypeUser              - client is authenticated as a user with login/password
  #  AuthTypeIoTDevice         - client is authenticated as an IoT device with certificate
  #  AuthTypeService           - client is authenticated as a service with certificate
  # The available capabilities depend on the auth type.
}


interface CapGatewayService  {
# CapGatewayService is the gateway service used to access Hub capabilities.

    newSession @0 (clientID :Text, sessionToken :Text) -> (session :Resolver.CapProvider);
    # Obtain a new gateway session for the given session token.
    # This fails with an error if the session token has already been used or is invalid.
    # A new sesion token can be obtained by the 'authXyz' methods of this service.
    # The capabilities available in the provided session depend on the token.

    authNoAuth @1 (clientID :Text) -> (sessionToken :Text);
    # Obtain a session token of an unauthenticated user.
    # clientID is the 'claimed' identity associated with the session
    #
    # Intended for an unprovisioned IoT device or unauthenticated client that need
    # capabilities available to unauthenticated users. For example, the provisioning capability.
    # This returns a token that can be used with newSession.

    authProxy @2 (clientID:Text, clientCertPEM:Text) -> (sessionToken :Text);
    # Obtain a new session token on behalf of a client with a certificate.
    #
    # Intended for a trusted proxy service that itself is authenticated using the connected
    # client certificate.
    #
    # clientID is that of the represented client and certPEM must be a valid certificate for that client.
    # This returns a token that can be used with newSession.

    authRefresh @3 (clientID:Text, sessionToken :Text) -> (sessionToken :Text);
    # Refresh the session token of a client with the given ID.
    # This returns a token that can be used with newSession.
    #
    # The default session token lifetime is 10 days and cannot be refreshed once expired.
    # The gateway can invalidate the previous session token.

    authWithCert @4 () -> (sessionToken:Text);
    # Authenticate using the connected client certificate.
    #
    # Intended for clients that use a client certificate authentication, such as devices
    # or services.
    # This returns a token that can be used with newSession.

    authWithPassword @5 (clientID:Text, password:Text) -> (sessionToken :Text);
    # Authenticate using a password.
    #
    # Intended for users with a login ID and password.
    # This returns a token that can be used with newSession.
}
