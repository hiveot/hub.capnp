# Cap'n proto definition for Hub gateway service
@0xdd3a962266ddd0e3;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");

using History = import "./History.capnp";
using Provisioning = import "./Provisioning.capnp";
using Directory = import "./Directory.capnp";

interface Gateway {
    # The gateway is the main entrypoint for access to the Hub
    # It provides capabilities to clients, based on their role

    login @0 (clientID:Text, password:Text) -> (cap :ClientCapabilities);
    # Login to the gateway and obtain the capabilities to use the gateway

    capRequestProvisioning @1 () -> (cap :Provisioning.CapRequestProvisioning);
    # obtain capability to request provisioning
    # intended for use by IoT devices
}

interface ClientCapabilities {
    # Result of login with available capabilities. Each is optional.
    # TBD: change to a dynamic list of capabilities based on the user role
    
    readDirectoryCapability @0 () -> (cap :Directory.CapReadDirectory);
    # Obtain capability to read from the Thing directory

    readHistoryCapability @1 (thingAddr :Text) -> (cap :History.CapReadHistory);
    # Obtain capability to iterate the history of a thing
}
