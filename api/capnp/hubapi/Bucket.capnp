@0x893d996fbc85a1c3;
using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub/api/go/hubapi");


struct KeyValueMap {
  # capnp doesn't support map types. It uses a struct with dynamic keys.
  # This compiles to an array in golang.
  entries @0 :List(Entry);
  struct Entry {
    key @0 :Text;
    value @1 :Data;
  }
}

struct BucketStoreInfo {
    # Bucket Store information of the bucket or the store

    dataSize @0 :Int64;
    # Size is the estimated disk space used by the store or bucket in bytes
    # -1 if not available

    engine @1 :Text;
    # Storage engine used, eg "kvbtree", "bolts", "pebble", "mongo"

    id @2 :Text;
    # ID holds the store or bucket identifier

    nrRecords @3 :Int64;
    # The number of records in the store or bucket
    # -1 if not available
}




interface CapBucketCursor  {
# CapBucketCursor provides the capability to iterate a bucket

    first @0 () -> (key :Text, value :Data, valid :Bool);
    # First positions the cursor at the first key in the ordered list
    # valid is false if the bucket is empty

    last @1 () -> (key :Text, value :Data, valid :Bool);
    # Last positions the cursor at the last key in the ordered list
    # valid is false if the bucket is empty

    next @2 () -> (key :Text, value :Data, valid :Bool);
    # Next moves the cursor to the next key from the current cursor
	# First() or Seek must have been called first.
    # valid is false if the iterator has reached the end and no valid value is returned.

    nextN @3 (steps :UInt32) -> (docs :KeyValueMap, itemsRemaining :Bool);
	# NextN moves the cursor to the next N places from the current cursor and return a map
	# with the N key-value pairs.
	# If the iterator reaches the end it returns the remaining items and itemsRemaining is false
	# If the cursor is already at the end, the resulting map is empty and itemsRemaining is also false.
	# Intended to speed up with batch iterations over rpc.

    prev @4 () -> (key :Text, value :Data, valid :Bool);
    # Prev moves the cursor to the previous key from the current cursor
	# Last() or Seek must have been called first.
	# valid is false if the iterator has reached the beginning and no valid value is returned.

    prevN @5 (steps :UInt32) -> (docs :KeyValueMap, itemsRemaining :Bool);
	# PrevN moves the cursor back N places from the current cursor and returns a map with
	# the N key-value pairs.
	# Intended to speed up with batch iterations over rpc.
	# If the iterator reaches the beginning it returns the remaining items and itemsRemaining is false
	# If the cursor is already at the beginning, the resulting map is empty and itemsRemaining is also false.

    seek @6 (searchKey :Text) -> (key :Text, value :Data, valid :Bool);
    # Seek positions the cursor at the given searchKey and corresponding value.
    # If the key is not found, the next key is returned.
	# valid is false if the iterator has reached the end and no valid value is returned.
}
