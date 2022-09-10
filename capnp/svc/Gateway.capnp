# Cap'n proto definition for Hub gateway service
@0xdd3a962266ddd0e3;

using Go = import "/go.capnp";
$Go.package("svc");
$Go.import("github.com/hiveot/hub.capnp/go/svc");

using EventHistory = import "./EventHistory.capnp";
using PropertyStore = import "./PropertyStore.capnp";
using Provisioning = import "./Provisioning.capnp";
using ThingDirectory = import "./ThingDirectory.capnp";

struct ClientCapabilities {
    # Result of login with available capabilities. Each is optional.
    
    directory @0 :ThingDirectory.ThingDirectory;
    #
    eventHistory @1 :EventHistory.EventHistoryStore;
    # 
    propertyStore @2 :PropertyStore.PropertyStore;
    #
    provisioning @3 :Provisioning.ProvisioningService;
    #
}


interface Gateway {
    # The gateway is the main entrypoint for access to the Hub
    # It provides capabilities to clients, based on their role

    login @0 (clientID:Text, password:Text) -> (clientCertPEM :Text);
    # Login to the gateway and obtain a client certificate for obtaining capabilities 

    capabilities @1 () -> (clientCap :ClientCapabilities);
    # Obtain the client capabilities based on the connection certificate
}

