# Cap'n proto definition for Thing data types
@0xbb31fb6e03b18e9a;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");

# Golang date format for 'created' time.
# Event timestamps must be formatted as YYYY-MM-DDTHH:MM:SS.sss-0700 where 0700 is the timezone.
const goISO8601Format :Text = "2006-01-02T15:04:05.000-0700";

struct ThingValue {
    # ThingValue holds events, actions or TD documents. Anything that comes from a Thing.
    # It contains contextual information related to the Thing such as its publisher (gatewayID)

    thingAddr @0 :Text;
    # Address of the thing owning the value.
    # Usually publisherID/thingID, where publisherID is the thingID of the publishing device.

    name @1 :Text;
    # Name of event or action as described in the thing TD
    # If the value holds a TD then this is 'td'

    valueJSON @2:Data;
    # Value, JSON encoded []byte array.

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
