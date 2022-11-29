# This is the Service interface that all services must implement in order for the gateway to be able to
# share out the capabilities of the service.
using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");
@0xaedf69d9667f51a8;

# types of clients
const clientTypeUnauthenticated  :Text = "noauth";
# ClientTypeUnauthenticated for clients without authentication

const clientTypeIotDevice :Text = "iotdevice";
# ClientTypeIotDevice for clients authenticated as IoT devices

const clientTypeUser :Text = "user";
# ClientTypeUser for clients authenticated as end-users

const clientTypeService :Text = "service";
# ClientTypeService for clients authenticated as Hub services

annotation clientType(method) :Text;
# annotation for client type allowed to use the method

struct CapabilityInfo  {
# CapabilityInfo provides information on a capabilities available through the gateway

	capabilityName @1 :Text;
	# Name of the capability. This is the method name as defined by the service.

	capabilityArgs @2 :List(Text);
	# list of arguments that is required

	clientTypes @3 :List(Text);
	# Type of clients that can use the capability. See ClientTypeXyz above

	serviceName @0 :Text;
	# Service name that is providing the capability.
}


interface CapHiveOTService {
# interface all Hive services implement

    getCapability @0 (clientID :Text, clientType :Text, capabilityName :Text, args :List(Text) ) -> (cap :Capability);
    # GetCapability obtains the capability with the given name.
    # This returns the client for that capability, or nil if the capability is not available. The result must be
    # cast to the corresponding interface.
    #
    # All capabilities must be released after use.
    #
    # clientID is the ID of the client, in case further ID related auth is needed.
   	# clientType is the type of authenticated client
   	# capabilityName is the name to retrieve
    # args is an array with arguments as per API with the same name

    listCapabilities @1 (clientType :Text) -> (infoList :List(CapabilityInfo));
    # List the available capabilities of this service
}
