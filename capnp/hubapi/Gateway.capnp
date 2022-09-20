# Cap'n proto definition for Hub gateway service
@0xdd3a962266ddd0e3;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");

using HistoryStore = import "./HistoryStore.capnp";
using PropertyStore = import "./PropertyStore.capnp";
using ProvisioningService = import "./ProvisioningService.capnp";
using DirectoryStore = import "./DirectoryStore.capnp";

struct ClientCapabilities {
    # Result of login with available capabilities. Each is optional.
    
    directoryStore @0 :DirectoryStore.DirectoryStore;
    #
    historyStore @1 :HistoryStore.HistoryStore;
    # 
    propertyStore @2 :PropertyStore.PropertyStore;
    #
    provisioningService @3 :ProvisioningService.ProvisioningService;
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
