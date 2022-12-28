# Cap'n proto definition for Hub gateway service
@0xdd3a962266ddd0e3;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");

using Resolver = import "Resolver.capnp";

struct ClientInfo {
# ClientInfo contains client info as seen by the gateway
# Intended for diagnostics and troubleshooting

  clientID @0 :Text;
  # ClientID that is connected. loginID, serviceID, or IoT device ID

  clientType @1 :Text;
  # ClientType identifies how the client is authenticated. See also the resolver
  #  ClientTypeUnauthenticated   - client is not authenticated
  #  ClientTypeUser              - client is authenticated as a user with login/password
  #  ClientTypeIoTDevice         - client is authenticated as an IoT device with certificate
  #  ClientTypeService           - client is authenticated as a service with certificate
  # The available capabilities depend on the client type.
}



interface CapGatewaySession {
    # The gateway session provides Hub capabilities to connected clients
    # Each client receives its own session which is used to track authentication state.
    #
    # Clients can authenticate with a signed client certificate or using the login method.

    listCapabilities @0 () -> (infoList :List(Resolver.CapabilityInfo));
	# ListCapabilities returns the list of capabilities available on the resolver

    registerCapabilities @1 (serviceID :Text, capInfo :List(Resolver.CapabilityInfo), provider :Resolver.CapProvider) -> ();
	# RegisterCapabilities is invoked by services that register capabilities.
	# provider is the callback interface of the service for obtaining the capabilities.

    login @2 (clientID:Text, password:Text) -> (authToken :Text, refreshToken :Text);
    # Login to the gateway as a user in order to get additional capabilities.
    # This returns an authToken and refreshToken that can be used with services that require
    # authentication.
    # If the authentication token has expired then call refresh.

    # User login to the gateway to use its capabilities. This is intended for end-users only

    ping @3 () -> (reply :ClientInfo);
    # ping the gateway, no authentication is required

    refresh @4 (refreshToken :Text) -> (authToken :Text, refreshToken :Text);
    # Refresh the token pair obtained at login
}
