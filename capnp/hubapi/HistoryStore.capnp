# Cap'n proto definition for Thing history storage service
@0xf1bd301f7c12caab;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");


struct ThingValue {
    # Data containing an event or action value of a thing
    
    name @0 :Text;
    # Name of event or action as described in the thing TD

    valueJSON @1:Text;
    # Value, JSON encoded

    created @2:Text;
    # Timestamp the value was created, in ISO8601 format. Default is 'now'
}


struct StoreInfo {
    # History Store information

    engine @0 :Text;
    # Storage engine used, eg "mongodb" or other

    nrActions @1 :Int64;
    # The number of actions in the store

    nrEvents @2 :Int64;
    # The number of events in the store

    uptime @3 :Int64;
    # Nr of sections the service is running
}

interface HistoryStore {
# History store service for storage of thing properties, events, and actions

  addAction @0 (thingID :Text, name:Text, valueJSON:Text, created:Text) -> ();
  # Add a Thing action with the given name and value to the action history
  # value is json encoded. Optionally include a 'created' ISO8601 timestamp
  #

  addEvent @1 (thingID :Text, name:Text, valueJSON:Text, created:Text) -> ();
  # Add an event to the event history

  getActionHistory @2 (thingID :Text, actionName :Text, after:Text, before:Text, limit:Int32) -> (values :List(ThingValue));
  # Return the history of a Thing action
  # before and after are timestamps in iso8601 format (YYYY-MM-DDTHH:MM:SS-TZ)

  getEventHistory @3 (thingID :Text, eventName :Text, after:Text, before:Text, limit:Int32) -> (values :List(ThingValue));
  # Return the history of a Thing event
  # before and after are timestamps in iso8601 format (YYYY-MM-DDTHH:MM:SS-TZ)

  info @4 () -> (statistics :StoreInfo);
  # Return storage information

}
