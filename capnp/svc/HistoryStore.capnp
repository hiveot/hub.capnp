# Cap'n proto definition for Thing history storage service
@0xf1bd301f7c12caab;

using Go = import "/go.capnp";
$Go.package("svc");
$Go.import("github.com/hiveot/hub.capnp/go/svc");


struct ThingValue {
    # Data containing an event, property or action value of a thing
    
    name @0 :Text;
    # Name of event, property or action as described in the thing TD

    valueJSON @1:Text;
    # Value, JSON encoded

    timestamp @2:Text;
    # Timestamp the value was created, in ISO8601 format. Default is 'now'
}


interface HistoryStore {
# History store service for storage of thing values

  getHistory @0 (thingID :Text, valueName :Text, after:time, before:time, limit:Int32) -> (values :List(ThingValue));
  # Return the history of a Thing value 

  addValue @1 (thingID :Text, value:ThingValue) -> ();
  # Add an event to the event history

}
