# Cap'n proto definition for resolver service
@0xf02d0b8fc1fe2004;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub/api/go/hubapi");

const resolverServiceName :Text = "resolver";

const defaultResolverAddress :Text = "/tmp/hiveot-resolver.socket";
# default socket path for the resolver


#--- client authentication types
const authTypeUnauthenticated  :Text = "unauthenticated";
# AuthTypeUnauthenticated for clients without authentication

const authTypeAdmin :Text = "admin";
# AuthTypeIotDevice for clients authenticated as IoT devices

const authTypeIotDevice :Text = "iotdevice";
# AuthTypeIotDevice for clients authenticated as IoT devices

const authTypeUser :Text = "user";
# AuthTypeUser for clients authenticated as end-users

const authTypeService :Text = "service";
# AuthTypeService for clients authenticated as Hub services
#---

struct CapabilityInfo  {
# CapabilityInfo provides information on a capabilities available through the gateway

	interfaceID @0 :UInt64;
	# Internal capnp ID of the interface that provides the capability.
	# This is typically the bootstrap interface of the service providing the method to get the capability.

	methodID @1 :UInt16;
	# Internal capnp method ID of the method that provides the capability.
    # This is the method index in the bootstrap interface above.

	interfaceName @2 :Text;
	# InterfaceName is the canonical name of the bootstrap interface providing the capability

	methodName @3 :Text;
    # MethodName is the name of the method in the bootstrap interface that provides the capability.
    # Method names must be unique and typically have the name of the capability interface.

	authTypes @4 :List(Text);
	# Required authentication types for using this capability. See AuthTypeXyz above.

	protocol @5 :Text;
	# The protocol to use; default is capnp. Services can also publish other protocols such
	# as rtsp and https. If a protocol is used that is not capnp, then dNetwork and dAddress are required.

	serviceID @6 :Text;
	# ServiceID is the instance ID of the service that is providing the capability.

	network @7 :Text;
	# optional direct access network; unix or tcp

	address @8 :Text;
	# Address is the connection address of the service implementing the interface and method
	# to obtain the capability.
	#  * leave empty to use the connection that provided this info (default)
	#  * unix domain sockets provide the socket path to dial into.
	#  * tcp networks provide the IP address:port, and optionally a path, depending on the protocol
}

interface CapResolverService extends (CapProvider) {
# CapResolverService provides an aggregated list of available capabilities from services
}


interface CapProvider {
# CapProvider provides capabilities from service providers.
# This interface is implemented by all service providers

	listCapabilities @0 (authType :Text) -> (infoList :List(CapabilityInfo));
	# ListCapabilities returns the list of provided capabilities
}
