# Cap'n proto definition for PubSub service
@0xf33c8b5943a21269;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub/api/go/hubapi");

using Bucket = import "./Bucket.capnp";
using Thing = import "./Thing.capnp";

const pubsubServiceName :Text = "pubsub";

const capNameDevicePubSub :Text = "capDevicePubSub";
const capNameServicePubSub :Text = "capServicePubSub";
const capNameUserPubSub :Text = "capUserPubSub";

const thingsPrefix :Text = "things";
# pubsub topic prefix used for Thing events and actions
const messageTypeAction :Text = "action";
# pubsub topic for action messages
const messageTypeEvent :Text = "event";
# pubsub topic for event messages

const eventNameProperties :Text = "properties";
# standardized event name containing a property key-value map
const eventNameTD :Text = "td";
# standardized event name containing a JSON serialized TD document

const actionNameConfig :Text = "configuration";
# standardized action name containing a key-value map for property configuration


interface CapPubSubService {
# CapPubSubService capabilities for publishing and subscribing to Thing messages

	capDevicePubSub @0 (deviceID :Text) -> (cap :CapDevicePubSub);
	# CapDevicePubSub provides the capability to pub/sub thing information as an IoT device.
	# The issuer must only provide this capability after authenticating the device.
	# The publisherID is the thingID of the device requesting the capability and
	# acts as the publisher of Things on this device.

	capServicePubSub @1 (serviceID :Text) -> (cap :CapServicePubSub);
	# CapServicePubSub provides the capability to pub/sub thing information as a service.
	# The issuer must only provide this capability after authenticating the service.
	# The publisherID is the thingID of the service requesting the capability and
	#  is the publisher of all Things published by the service.

	capUserPubSub @2 (userID :Text) -> (cap :CapUserPubSub);
	# CapUserPubSub provides the capability to pub/sub thing information as an end-user.
	# The issuer must only provide this capability after authenticating the user.
	# The userID is the loginID of the user requesting the capability.
}


interface CapDevicePubSub {
# CapDevicePubSub capabilities available to an IoT device

	pubEvent @0 (thingID :Text, eventID :Text, value :Data) -> ();
    # PubEvent publishes the given thing event. The payload is an event value as per TD.
	# This will combine the device's publisherID with the given thingID in the publication address
	#  thingID of the Thing whose event is published
	#  eventID is the key of the event affordance in the Thing's event map
	#  value is the serialized event value, or nil if the event has no value

	subAction  @1 (thingID :Text, actionID :Text, handler :CapSubscriptionHandler) -> ();
	# SubAction subscribes to actions aimed at things published by this device.
	# To request configuration changes use the "properties" action ID.
	#  thingID is the thing to subscribe for, or "" to subscribe to all things of this gateway
	#  actionID is the ID in the actions map, or "" to subscribe to all actions including 'configurationActionName'
	#   for configuration requests.
	#  handler will be invoked when an action request is received
}

interface CapServicePubSub extends(CapDevicePubSub, CapUserPubSub) {
# CapServicePubSub is the publish/subscribe capability available to Hub services.
# Hub services have IoT device capabilities and consumer capabilities as publishers of their own service and can
# subscribe similar to consumers. In addition to all events, actions and TDs.

	subActions @0 (publisherID :Text, thingID :Text, actionID :Text, handler :CapSubscriptionHandler) -> ();
	# SubActions subscribes to actions aimed at things from any publisher.
	#
	# This is intended for services that track actions aimed at other devices or services,
	# Possibly a logging, monitoring or automation service.
	#
	#  publisherID is the ID of the publisher that is receiving the actions, or "" for all publishers.
	#  thingID is the ID of the Thing whose action to subscribe to or "" for all things of the selected publisher.
	#  actionID or "" to subscribe to all actions
	#  handler is a callback invoked when actions are received

	subEvents @1 (publisherID :Text, thingID :Text, eventID :Text, handler :CapSubscriptionHandler) -> ();
	# subEvents subscribes to events from things from any publisher.
	#
	# It is not recommended to subscribe to all events of all things of all publishers unless
	# absolutely needed for the task at hand, as this can generate a lot of calls.
	#
	#  publisherID is the ID of the publisher that is publishing the event.
	#  thingID is the ID of the Thing whose event to subscribe to or "" for all things published by the publisher.
	#  eventID or "" to subscribe to all events.
	#  handler is a callback invoked when actions are received
}


interface CapUserPubSub {
# CapUserPubSub is the publish/subscribe capability available to Hub end-users.

	pubAction @0 (publisherID :Text, thingID :Text, actionID :Text, value :Data) -> ();
    # PubAction publishes an action request for a Thing.
    #
	# Authorization will only allow actions to be published for things that are in the same group as
	# the user and for which the user has the operator or manager role.
	# The manager role allows sending the hubapi.ActionNameConfiguration action.
	#
	#  publisherID is the ID of the device or service that is publishing the thing
	#  thingID is the ID of the Thing whose action is being requested
	#  actionID is the ID as defined in the Thing's TD or configurationActionName for configuration of writable properties
	#  value is the JSON encoded value of the action

	subEvent @1 (publisherID :Text, thingID :Text, eventID :Text, handler :CapSubscriptionHandler) -> ();
	# SubEvent subscribes to events from a thing
	#
	# It is not allowed to subscribe to all events of all things of all publishers.
	# A thingID or an eventID must be provided or this will return an error.
	#
	#  publisherID is the ID of the device or service that is publishing the thing event.
	#  thingID is the ID of the Thing whose event is published.
	#  eventID of the event. Use "" to subscribe to all events of the things.
}



interface CapSubscriptionHandler {
# SubscriptionHandler is the callback interface for subscriptions

   handleValue @0 (value :Thing.ThingValue) -> ();
}
