# Cap'n proto definition for storage of thing property values
@0x907dec5c486f8a26;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");


struct PropValue {
# Property values as received by the service

    name @0: Text;
    # Name of the property

    valueJSON @1: Text;
    # Value of the property, JSON encoded

    timestamp @2: Text;
    # Timestamp the property value was updated
}



interface PropertyStore {
# Service for storage of property values

  getThingValues @0 (thingID:Text) -> (values:List(PropValue));
  # Return the most recent property values of a Thing

  getPropertyHistory @1 (thingID:Text, name:Text, limit:Int32) -> (values :List(PropValue));
  # Return the history of a single property value of a Thing

  updatePropertyValue @2 (thingID:Text, value:PropValue ) -> ();
  # Update a single property value of a thing

  updatePropertyValues @3 (thingID:Text, propValues:List(PropValue)) -> ();
  # Update multiple property values of a thing
}
