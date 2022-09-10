# Cap'n proto definition for Thing directory service
@0xc8da54a8b024bd49;

using Go = import "/go.capnp";
$Go.package("svc");
$Go.import("github.com/hiveot/hub.capnp/go/svc");

using ThingDescription = import "../thing/ThingDescription.capnp";


interface ThingDirectory {
  # Thing Directory service

  getTD @0 (id :Text) -> (td :ThingDescription.TD);
  # Return the TD with the given Thing ID

  queryTDs @1 (jsonPath :Text, limit:Int32, offset:Int32) -> (tds :List(ThingDescription.TD));
  # Query for TD's using JSONpath on the TD content
  # See 'docs/query-tds.md' for examples 
  
  listTDs @2 (limit:Int32, offset:Int32) -> (tds :List(ThingDescription.TD));
  # List all TD's
  
  updateTD @3 (td :ThingDescription.TD) -> ();
  # Update the TD document in the directory
  # If the TD with the given ID doesn't exist it will be added.
}
