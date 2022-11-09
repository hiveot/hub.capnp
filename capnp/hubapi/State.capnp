# Cap'n proto definition for state store
@0x9a80401eba6f7fe3;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");



struct KeyValueMap {
  # capnp doesn't have map types. It uses a struct with dynamic keys.
  # This compiles to an array in golang.
  entries @0 :List(Entry);
  struct Entry {
    key @0 :Text;
    value @1 :Data;
  }
}

interface CapState {
  # State storage

    capClientState @0 (clientID :Text, appID :Text) -> (cap :CapClientState);
    # Get the capability to store state for a client application
}


interface CapBucketCursor  {
# CapBucketCursor provides the capability to iterate a bucket

	first @0 () -> (key :Text, value :Data);
	# First positions the cursor at the first key in the ordered list

	last @1 () -> (key :Text, value :Data);
	# Last positions the cursor at the last key in the ordered list

	next @2 () -> (key :Text, value :Data);
	# Next moves the cursor to the next key from the current cursor

	prev @3 () -> (key :Text, value :Data);
	# Prev moves the cursor to the previous key from the current cursor

	seek @4 (searchKey :Text) -> (key :Text, value :Data);
	# Seek positions the cursor at the given searchKey and corresponding value.
	# If the key is not found, the next key is returned.
	# cursor.Close must be invoked after use in order to close any read transactions.
}

interface CapClientState {
# Capability for reading and writing state values

  delete @0 (key :Text) -> ();
  # Delete removes the key-value pair from the state store

  get @1 (key :Text) -> (value :Data);
  # Get returns the document for the given key
  # Returns an error if the key doesn't exist  # Get state value for key

  getMultiple @2 (keys :List(Text)) -> (docs :KeyValueMap);
  # Get returns the document for the given key
  # Returns an error if the key doesn't exist  # Get state value for key

  cursor @3 () -> (cap :CapBucketCursor);
  # Cursor returns the capability to iterate the client bucket

  set @4 (key :Text, value :Data) -> ();
  # Set updates a document with the given key in the store

  setMultiple @5 (docs :KeyValueMap) -> ();
  # SetMultiple sets multiple documents in a batch update

}
