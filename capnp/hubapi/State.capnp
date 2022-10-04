# Cap'n proto definition for state store
@0x9a80401eba6f7fe3;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");


interface CapState {
  # State storage

    capClientState @0 (clientID :Text, appID :Text) -> (cap :CapClientState);
    # Get the capability to store state for a client application
}


interface CapClientState {
# Capability for reading and writing state values

  get @0 (key :Text) -> (value :Text);
  # Get state value for key

  set @1 (key :Text, value :Text) -> ();
  # Set value of key
}
