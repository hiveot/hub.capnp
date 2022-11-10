@0x893d996fbc85a1c3;
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

interface CapBucketCursor  {
# CapBucketCursor provides the capability to iterate a bucket

    first @0 () -> (key :Text, value :Data);
    # First positions the cursor at the first key in the ordered list

    last @1 () -> (key :Text, value :Data);
    # Last positions the cursor at the last key in the ordered list

    next @2 () -> (key :Text, value :Data);
    # Next moves the cursor to the next key from the current cursor

    nextN @3 (steps :UInt32) -> (docs :KeyValueMap, endReached :Bool);
    # NextN moves the cursor N steps from the current cursor and returns the collected KV pairs

    prev @4 () -> (key :Text, value :Data);
    # Prev moves the cursor to the previous key from the current cursor

    prevN @5 (steps :UInt32) -> (docs :KeyValueMap, startReached :Bool);
    # PrevN moves the cursor N steps backwards from the current cursor and returns the collected KV pairs

    seek @6 (searchKey :Text) -> (key :Text, value :Data);
    # Seek positions the cursor at the given searchKey and corresponding value.
    # If the key is not found, the next key is returned.
    # cursor.Close must be invoked after use in order to close any read transactions.
}
