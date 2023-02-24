# Cap'n proto definition for Thing history storage service
@0xf1bd301f7c12caab;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub/api/go/hubapi");
using Thing = import "./Thing.capnp";
using Bucket = import "./Bucket.capnp";


const historyServiceName :Text = "history";

const capNameAddHistory :Text = "capAddHistory";
const capNameManageRetention :Text = "capManageRetention";
const capNameReadHistory :Text = "capReadHistory";


struct EventRetention {
# Event retention rule

   name @0 :Text;
   # event or action name to retain

   publishers @1 :List(Text);
   # optional, only accept this event from these publishers, default is all

   things @2 :List(Text);
   # optional, only accept this event from these things, default is all

   exclude @3 :List(Text);
   # optional, exclude this event from these things

   retentionDays @4 :Int32;
   # todo: remove events older than retentionDays. 0 is indefinitely
}



interface CapHistoryService {
# Available History store capabilities

  capAddHistory @0 (clientID :Text, ignoreRetention :Bool) -> (cap :CapAddHistory);
  # capAddHistory provides the capability to add to the history of any Thing.
  # This capability should only be provided to trusted services that capture events from multiple
  # sources and can verify their authenticity.
  #  clientID is the client requesting the capability
  #  ignoreRetention means that retention rules are not applied when adding the event to history

  capManageRetention @1 (clientID :Text) -> (cap :CapManageRetention);
  # capManageRetention provides the capability to change the event retention rules
  # Intended for use by administrators.
  #  clientID is the client requesting the capability

  capReadHistory @2 (clientID :Text, publisherID :Text, thingID :Text) -> (cap :CapReadHistory);
  # CapReadHistory provides the capability to iterate history.
  # This returns an iterator for the history.
  # The cursor key is the timestamp in ISO8601 in msec, eg YYYY-MM-DDTHH:MM:SS.sss-TZ
  #  the cursor value is the event or action
  # Values added after creating the cursor might not be included, depending on the
  # underlying store.
  # This capability can be provided to anyone who has read access to the thing.
  #
  #  clientID is the client requesting the capability
  #  publisherID to restrict reading to
  #  thingID to restrict reading to
}



interface CapAddHistory {
# Capability to add to a Thing's history

  addAction @0 (tv :Thing.ThingValue) -> ();
  # Add a Thing action with the given name and value to the action history
  # value is json encoded. Optionally include a 'created' ISO8601 timestamp

  addEvent @1 (tv :Thing.ThingValue) -> ();
  # Add an event to the event history

  addEvents @2 (tv :List(Thing.ThingValue)) -> ();
  # Bulk add events to the event history
}


interface CapManageRetention {
# Capability to manage the event retention rules

  getEvents @0 () -> (retList :List(EventRetention));
  # Return the list of events to retain

  getEventRetention @1 (name :Text) -> (ret :EventRetention);
  # Add an event to the event history

  removeEventRetention @2 (name :Text) -> ();
  # Remove existing retention rule, if it exists.

  setEventRetention @3 (ret :EventRetention) -> ();
  # Bulk add events to the event history

  testEvent @4 (tv :Thing.ThingValue) -> (retained :Bool);
  # Test if the given event passes the retention rules for adding to the history

}


interface CapReadHistory {
# CapReadHistory defines the capability to read information from a thing

	getEventHistory @0 (name :Text) -> (cursor :CapHistoryCursor);
	# GetEventHistory returns a cursor to iterate the history of a thing's event
	# name is the event or action to filter on. Use "" to iterate all events/action of the thing

	getProperties @1 (names :List(Text)) -> (valueList :List(Thing.ThingValue));
	# GetProperties returns the most recent property and event values of the Thing

	info @2 () -> (info :Bucket.BucketStoreInfo);
	# info() returns the storage information of the Thing
}

interface CapHistoryCursor {
# CapHistoryCursor is a cursor to iterate the Thing event and action history
# This is a bucket cursor that converts converts the data to ThingValue types.
# Use Seek to find the start of the range and NextN to read batches of values

	first @0 () -> (tv :Thing.ThingValue, valid :Bool);
    # First return the oldest value in the history
	# Returns nil if the store is empty

	last @1 () -> (tv :Thing.ThingValue, valid :Bool);
	# Last returns the latest value in the history
	# Returns nil if the store is empty

	next @2 () -> (tv :Thing.ThingValue, valid :Bool);
    # Next returns the next value in the history
	# Returns nil when trying to read past the last value

	nextN @3 (steps :UInt32) -> (batch :List(Thing.ThingValue), valid :Bool);
	# NextN returns a batch of next history values
	# Returns empty list when trying to read past the last value

	prev @4 () -> (tv :Thing.ThingValue, valid :Bool);
	# Prev returns the previous value in history
	# Returns nil when trying to read before the first value

	prevN @5 (steps :UInt32) -> (batch :List(Thing.ThingValue), valid :Bool);
	# PrevN returns a batch of previous history values
	# Returns empty list when trying to read before the first value

	seek @6 (isoTimestamp :Text) -> (tv :Thing.ThingValue, valid :Bool);
	# Seek the starting point for iterating the history
	# This returns the value at timestamp or next closest if it doesn't exist
    # The timestamp is in iso8601 format (YYYY-MM-DDTHH:MM:SS-TZ)
	# Returns empty list when there are no values at or past the given timestamp

}
