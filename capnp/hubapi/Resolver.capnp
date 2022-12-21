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

	capabilityName @0 :Text;
	# Name of the capability. This is the capnp interface name as defined by the service.

	capabilityArgs @1 :List(Text);
	# list of argument names that are required.

	clientTypes @2 :List(Text);
	# Type of clients that can use the capability. See ClientTypeXyz above

	protocol @6 :Text;
	# The protocol to use; default is capnp. Services can also publish other protocols such
	# as rtsp and https. If a protocol is used that is not capnp, then dNetwork and dAddress are required.

	serviceID @3 :Text;
	# ServiceID is the instance ID of the service that is providing the capability.

	dNetwork @4 :Text;
	# optional direct access network; unix or tcp

	dAddress @5 :Text;
	# optional direct access address; socket path or network address
}

interface CapResolverSession  {
# CapResolverSession is a client of the resolver service using to access and register capabilities.

	getCapability @0 (clientID :Text, clientType :Text, capName :Text, args :List(Text)) -> (capability :Capability);
	# GetCapability returns a capability that was registered with the resolver.
	#
    # Peers MUST provide the clientID and type of the client requesting the capability. Clients that
    # request capabilities on behalf of others must be authenticated and the clientType must be provided.
    #
    # The capability that is returned must be released after use by the remote peer.
	#
	#  clientID is the ID of the authenticated client requesting the capability.
	#  clientType is the verified type of client.
	#  capName is the capability name as published in RegisterCapabilities.
	#  args is optional list of arguments when needed for the requested capability.
	#
	# This returns the capability or an error if the capability is not available.
	# The error types are ResolveDenied when the client is not allowed access to the capability, and ResolveUnavailable
	# if the capability is not available.

	listCapabilities @1 () -> (infoList :List(CapabilityInfo));
	# ListCapabilities returns the list of capabilities available on the resolver

    registerCapabilities @2 (serviceID :Text, capInfo :List(CapabilityInfo), provider :CapProvider) -> ();
	# RegisterCapabilities is invoked by services that register capabilities.
	# provider is the callback interface of the service for obtaining the capabilities.
}

interface CapProvider {
# CapProvider provides capabilities from service providers.
# This is the callback in to RegisterCapabilities to provide the service capabilities

	getCapability @0 (clientID :Text, clientType :Text, capabilityName :Text, args :List(Text)) -> (capability :Capability);
	# GetCapability returns the requested capability.
	#
    # Peers MUST provide the clientID and type of the client requesting the capability. Clients that
    # request capabilities on behalf of others must be authenticated and the clientType must be provided.
    #
    # The capability that is returned must be released after use by the remote peer.
	#
	#  name is the capability name as published in RegisterCapabilities.
	#  clientID is the ID of the authenticated client requesting the capability.
	#  clientType is the verified type of client.
	#  args is optional list of arguments when needed for the requested capability.
	#
	# This returns the capability or an error if the capability is not available on the provider.
	# The error types are ResolveDenied when the client is not allowed access to the capability, and ResolveUnavailable
	# if the capability is not available.

	listCapabilities @1 () -> (infoList :List(CapabilityInfo));
	# ListCapabilities returns the list of provided capabilities
}
