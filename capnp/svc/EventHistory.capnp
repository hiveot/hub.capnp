# Cap'n proto definition for Thing event storage service
@0xf1bd301f7c12caab;

using Go = import "/go.capnp";
$Go.package("svc");
$Go.import("github.com/hiveot/hub.capnp/go/svc");


struct ThingEvent {
    # Data containing an event
    
    name @0 :Text;
    # Name of event as described in the thing TD

    valueJSON @1:Text;
    # Value of the event, JSON encoded

    timestamp @2:Text;
    # Timestamp of the event in ISO8601 format. Default is 'now'
}


interface EventHistoryStore {
# History store service for storage of thing events

  getEventHistory @0 (thingID :Text) -> (events :List(ThingEvent));
  # Return the history of a Thing event

  addEvent @1 (thingID :Text, event:ThingEvent) -> ();
  # Add an event to the event history

}
