# Cap'n proto definition for state store
@0x9a80401eba6f7fe3;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub/api/go/hubapi");

using Bucket = import "./Bucket.capnp";

const stateServiceName :Text = "state";

const capNameClientState :Text = "capClientState";

interface CapState {
  # State storage

    capClientState @0 (clientID :Text, appID :Text) -> (cap :CapClientState);
    # Get the capability to store state for a client application
}



interface CapClientState {
# Capability for reading and writing state values

  delete @0 (key :Text) -> ();
  # Delete removes the key-value pair from the state store

  get @1 (key :Text) -> (value :Data);
  # Get returns the document for the given key
  # Returns an error if the key doesn't exist  # Get state value for key

  getMultiple @2 (keys :List(Text)) -> (docs :Bucket.KeyValueMap);
  # Get returns the document for the given key
  # Returns an error if the key doesn't exist  # Get state value for key

  cursor @3 () -> (cap :Bucket.CapBucketCursor);
  # Cursor returns the capability to iterate the client bucket

  set @4 (key :Text, value :Data) -> ();
  # Set updates a document with the given key in the store

  setMultiple @5 (docs :Bucket.KeyValueMap) -> ();
  # SetMultiple sets multiple documents in a batch update

}
