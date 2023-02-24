# Cap'n proto definition for Thing directory store
@0xc8da54a8b024bd49;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub/api/go/hubapi");

using Bucket = import "./Bucket.capnp";
using Thing = import "./Thing.capnp";

const directoryServiceName :Text = "directory";

const capNameReadDirectory :Text = "capReadDirectory";
const capNameUpdateDirectory :Text = "capUpdateDirectory";

interface CapDirectoryService {
  # Available Thing directory capabilities

  capReadDirectory @0 (clientID :Text) -> (cap :CapReadDirectory);
  # Capabilities to read the directory

  capUpdateDirectory @1 (clientID :Text) -> (cap :CapUpdateDirectory);
  # Capabilities to update the directory
}

interface CapDirectoryCursor {
# Cursor to iterate the directory. Obtained from CapReadDirectory.

	first @0 () -> (tv :Thing.ThingValue, valid :Bool);
    # First return the oldest value in the directory
	# Returns nil if the store is empty

	next @1 () -> (tv :Thing.ThingValue, valid :Bool);
    # Next returns the next value in the directory
	# Returns nil when trying to read past the last value

	nextN @2 (steps :UInt32) -> (batch :List(Thing.ThingValue), valid :Bool);
	# NextN returns the next batch of directory entries
	# Returns empty list when trying to read past the last value
}

interface CapReadDirectory {
# Capability to read from the directory

  cursor @0 () -> (cursor :CapDirectoryCursor);
  # Cursor returns an iterator for TDs 

  getTD @1 (publisherID :Text, thingID :Text) -> (tv :Thing.ThingValue);
  # Returns a ThingValue containing a TD document in JSON format.
  #  publisherID is the ID of the device publishing the Thing
  #  thingID is the ID of the thing.

}


interface CapUpdateDirectory {
# Capability to update the directory

  removeTD @0 (publisherID :Text, thingID :Text) -> ();
  # Remove the TD document with the given publisher/thing from the directory

  updateTD @1 (publisherID :Text, thingID :Text, tdDoc :Data) -> ();
  # Update the TD document in the directory.
  # If the TD doesn't exist it will be added.
}
