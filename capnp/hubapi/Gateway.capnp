# Cap'n proto definition for Hub gateway service
@0xdd3a962266ddd0e3;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");

using Service = import "Service.capnp";


#interface Service {};
# generic interface of a service

struct CapabilityInfo {
    service @0 :Text;
    name @1 :Text;
    clientType @2 :List(Text);
}


struct GatewayInfo {
# GatewayInfo describes the gateway's capabilities and capacity

    capabilities @0 :List(CapabilityInfo);
    url @1 :Text;
    latency @2 :Int32;
}

#interface Capability(T) {
# Return the capability of the given type
#}

interface CapGatewayService extends (Service.HiveService) {
    # The gateway is the main entrypoint for access to the Hub
    # It provides capabilities to clients, based on their role

    #login @0 (clientID:Text, password:Text) -> (cap :ClientCapabilities);
    # Login to the gateway and obtain the capabilities to use the gateway

    getCapability @0 (clientType :Text, service :Text) -> (cap :Service.HiveService);
	# GetCapability obtains the capability of the service with the given name, if available
	#
	# This returns the client for that capability, or nil if the capability is not available. The result must be
	# cast to the corresponding interface. (TBD: can we use generics?)
	#
	# All capabilities must be released after use. If the gateway capability is released or disconnected, then
	# all capabilities obtained via the gateway are also released.
	#
	# The capabilities that are available depend on how the client is authenticated at and whether it is
	# a device, service or end-user client.
	#
	#  clientType is the type of authenticated client
	#  service is the name of the service providing capabilities

    getGatewayInfo @1 () -> (info :GatewayInfo);

    ping @2 () -> (response :Text);

}
