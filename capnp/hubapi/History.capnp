# Cap'n proto definition for Thing history storage service
@0xf1bd301f7c12caab;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");

# Date format madness at work here.
# Event timestamps must be formatted as YYYY-MM-DDTHH:MM:SS.sss-0700 where 0700 is the timezone.
# in golang:
const goISO8601Format :Text = "2006-01-02T15:04:05.000-0700";

struct ThingValue {
    # Data containing an event or action value of a thing
    thingID @0 :Text;
    # ID of the thing owning the value

    name @1 :Text;
    # Name of event or action as described in the thing TD

    valueJSON @2:Text;
    # Value, JSON encoded

    created @3:Text;
    # Timestamp the value was created, in ISO8601 format. Default is 'now'
}

struct ThingValueMap {
  # capnp doesn't have map types. It uses a struct with dynamic keys.
  # This compiles to an array in golang.
  entries @0 :List(Entry);
  struct Entry {
    key @0 :Text;
    value @1 :ThingValue;
  }
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
    # Nr of seconds the service is running
}

interface CapHistory {
# Available History store capabilities

  capReadHistory @0 () -> (cap :CapReadHistory);
  # Capabilities to read the event and action history

  capUpdateHistory @1 () -> (cap :CapUpdateHistory);
  # Capabilities to update the Thing history
  # TBD. Limit to a specific Thing ID or publisher when used by a device directly

}

interface CapReadHistory {
# Capability to read from the history store

  getActionHistory @0 (thingID :Text, actionName :Text, after:Text, before:Text, limit:Int32) -> (values :List(ThingValue));
  # Return the history of a Thing action
  # before and after are timestamps in iso8601 format (YYYY-MM-DDTHH:MM:SS-TZ)

  getEventHistory @1 (thingID :Text, eventName :Text, after:Text, before:Text, limit:Int32) -> (values :List(ThingValue));
  # Return the history of a Thing event
  # before and after are timestamps in iso8601 format (YYYY-MM-DDTHH:MM:SS-TZ)
  # if before is provided, after must be provided as well

  getLatestEvents @2 (thingID :Text) -> (thingValueMap :ThingValueMap);
  # Return a map with the most recent event values of a Thing

  info @3 () -> (statistics :StoreInfo);
  # Return storage information
}

interface CapUpdateHistory {
# Capability to update the history store

  addAction @0 (actionValue :ThingValue) -> ();
  # Add a Thing action with the given name and value to the action history
  # value is json encoded. Optionally include a 'created' ISO8601 timestamp


  addEvent @1 (eventValue :ThingValue) -> ();
  # Add an event to the event history

  addEvents @2 (eventValues :List(ThingValue)) -> ();
  # Bulk add events to the event history

}
