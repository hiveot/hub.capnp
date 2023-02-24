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
  # The available capabilities depend on the client type.
}



interface CapGatewaySession  {

	listCapabilities @0 () -> (infoList :List(Resolver.CapabilityInfo));
	# ListCapabilities returns the list of provided capabilities
	# the result depends on the client's authentication type

    login @1 (clientID:Text, password:Text) -> (authToken :Text, refreshToken :Text);
    # Login to the gateway as a user in order to get additional capabilities.
    # This returns an authToken and refreshToken that can be used with services that require
    # authentication.
    # If the authentication token has expired then call refresh.

    # User login to the gateway to use its capabilities. This is intended for end-users only

    ping @2 () -> (reply :ClientInfo);
    # ping the gateway, no authentication is required

    refresh @3 (clientID:Text, refreshToken :Text) -> (authToken :Text, refreshToken :Text);
    # Refresh the token pair obtained with login
}
