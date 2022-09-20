# Cap'n proto definition for Thing directory store
@0xc8da54a8b024bd49;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");


interface DirectoryStore {
  # Thing Directory store

  getTD @0 (thingID :Text) -> (tdJson :Text);
  # Return the TD with the given Thing ID in JSON format

  queryTDs @1 (jsonPath :Text, limit:Int32, offset:Int32) -> (tds :List(Text));
  # Query for TD's using JSONpath on the TD content
  # See 'docs/query-tds.md' for examples 
  
  listTDs @2 (limit:Int32, offset:Int32) -> (tds :List(Text));
  # List all TD's
  
  updateTD @3 (thingID :Text, tdDoc :Text) -> ();
  # Update the TD document in the directory
  # If the TD with the given ID doesn't exist it will be added.
}
