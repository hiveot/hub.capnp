# Cap'n proto definition for Hub gateway service
@0xdd3a962266ddd0e3;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");

using Service = import "Service.capnp";


interface CapGatewayService   {
    # The gateway is the main entrypoint for access to the Hub.
    # It provides capabilities to clients, based on their role.
    # Clients can authenticate with a signed client certificate or using the login method.

	getCapability @0 (clientID :Text, clientType :Text, capabilityName :Text, args :List(Text)) -> (capability :Capability);
	# GetCapability returns the capnp capability of the given name
	# The client login determines what capabilities are available.
	#  clientID of the client requesting the capability
	#  args is optional in case the capability has additional parameters

	listCapabilities @1 (clientType :Text) -> (infoList :List(Service.CapabilityInfo));
	# ListCapabilities returns the aggregated list of capabilities from all connected services
	# This list is reduced to capabilities available to the client based on its authentication method

    login @2 (clientID:Text, password:Text) -> (success :Bool);
    # User login to the gateway to use its capabilities. This is intended for end-users only

    ping @3 () -> (response :Text);
    # ping the gateway, no authentication is required
}
