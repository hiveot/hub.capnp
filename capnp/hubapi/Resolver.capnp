# Cap'n proto definition for resolver service
@0xf02d0b8fc1fe2004;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");


annotation clientType(method) :Text;
# annotation for client type allowed to use the method

const defaultResolverAddress :Text = "/tmp/hive-resolver.socket";
# socket path for the default resolver


#--- types of clients
const clientTypeUnauthenticated  :Text = "noauth";
# ClientTypeUnauthenticated for clients without authentication

const clientTypeIotDevice :Text = "iotdevice";
# ClientTypeIotDevice for clients authenticated as IoT devices

const clientTypeUser :Text = "user";
# ClientTypeUser for clients authenticated as end-users

const clientTypeService :Text = "service";
# ClientTypeService for clients authenticated as Hub services
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

	clientTypes @4 :List(Text);
	# Type of clients that can use the capability. See ClientTypeXyz above

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

interface CapResolverService  {
# CapResolverService provides an aggregated list of available capabilities from services

	listCapabilities @0 (clientType :Text) -> (infoList :List(CapabilityInfo));
	# ListCapabilities returns the list of capabilities available on the resolver
}


interface CapProvider {
# CapProvider provides capabilities from service providers.
# This interface is implemented by all service providers

	listCapabilities @0 () -> (infoList :List(CapabilityInfo));
	# ListCapabilities returns the list of provided capabilities
}
