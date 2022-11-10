# Cap'n proto definition for Thing directory store
@0xc8da54a8b024bd49;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");

using Bucket = import "./Bucket.capnp";

interface CapListCallback {
# callback interface  to be implemented by callers of ListTDcb

   handler @0 (tds :List(Text), isLast :Bool) -> ();
   # handler is a method that receives a batch of TD documents
   #  tds is a list of TD documents in JSON format
   #  isLast is true if this is the last batch to be received
}


interface CapDirectory {
  # Available Thing directory capabilities

  capReadDirectory @0 () -> (cap :CapReadDirectory);
  # Capabilities to read the directory

  capUpdateDirectory @1 () -> (cap :CapUpdateDirectory);
  # Capabilities to update the directory
}

interface CapReadDirectory {
# Capability to read from the directory

  cursor @0 () -> (cursor :Bucket.CapBucketCursor);
  # Cursor returns an iterator for TD documents

  getTD @1 (thingID :Text) -> (tdJson :Data);
  # Return the JSON serialized TD with the given Thing ID in JSON format

}


interface CapUpdateDirectory {
# Capability to update the directory

  removeTD @0 (thingID :Text) -> ();
  # Remove the TD document in the directory

  updateTD @1 (thingID :Text, tdDoc :Data) -> ();
  # Update the TD document in the directory
  # If the TD with the given ID doesn't exist it will be added.
  # tdDoc is the JSON serialized TD document
}
