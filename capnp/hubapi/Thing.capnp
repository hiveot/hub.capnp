# Cap'n proto definition for Thing data types
@0xbb31fb6e03b18e9a;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");

# Golang date format for 'created' time.
# Event timestamps must be formatted as YYYY-MM-DDTHH:MM:SS.sss-0700 where 0700 is the timezone.
const goISO8601Format :Text = "2006-01-02T15:04:05.000-0700";

struct ThingValue {
    # Data containing an event or action value of a thing

    thingID @0 :Text;
    # ID of the thing owning the value

    name @1 :Text;
    # Name of event or action as described in the thing TD

    valueJSON @2:Data;
    # Value, JSON encoded []byte array

    created @3:Text;
    # Timestamp the value was created, in ISO8601 format (see above).
}

struct ThingValueMap {
  # capnp doesn't have map types. It uses a struct with dynamic keys.
  # This compiles to an array in golang which the (de)serializer turns back into a map.

  entries @0 :List(Entry);
  struct Entry {
    key @0 :Text;
    value @1 :ThingValue;
  }
}
