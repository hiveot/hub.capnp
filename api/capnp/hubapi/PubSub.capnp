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
	# This will combine the thingID with the device's thingID to publish it under the thing address
	#  thingID of the Thing whose event is published
	#  eventID is the key of the event affordance in the Thing's event map
	#  value is the serialized event value, or nil if the event has no value

	pubProperties  @1 (thingID :Text, props :Bucket.KeyValueMap) -> ();
	# PubProperties creates a topic and publishes properties of a thing.
	# This will combine the thingID with the device's thingID to publish it under the thing address
	#  thingID of the Thing whose event is published (not the thing address)
	#  The props is a map of property name-value pairs.

	pubTD  @2 (thingID :Text, tdDoc :Data) -> ();
	# PubTD publishes the given thing TD. The payload is a serialized TD document.
	# This will combine the publisher thingID with the device's thingID as the thing address
	#  thingID of the Thing whose event is published (not the thing address)

	subAction  @3 (thingID :Text, actionID :Text, handler :CapSubscriptionHandler) -> ();
	# SubAction creates a topic and registers a listener for actions to things with this gateway.
	# This supports receiving queued messages for this gateway since it last disconnected.
	#  thingID is the thing to subscribe for, or "" to subscribe to all things of this gateway
	#  actionID is the ID in the actions map, or "" to subscribe to all actions
	#  handler will be invoked when an action is received for this device
}

interface CapServicePubSub extends(CapDevicePubSub, CapUserPubSub) {
# CapServicePubSub is the publish/subscribe capability available to Hub services.
# Hub services have IoT device capabilities and consumer capabilities as publishers of their own service and can
# subscribe similar to consumers. In addition to all events, actions and TDs.

	subActions @0 (publisherID :Text, thingID :Text, actionID :Text, handler :CapSubscriptionHandler) -> ();
	# SubActions subscribes to actions aimed at things.
	# Services can subscribe to other actions for logging, automation and other use-cases.
	# For subscribing to service directed actions, use SubAction.
	#
	#  publisherID is the ID of the publisher that is receiving the actions.
	#   normally that would be the serviceID but services can also subscribe
	#   to actions send to other things.
	#  thingID is the ID of the Thing whose action to subscribe to or "" for
	#   all things published by the publisher.
	#  actionID or "" to subscribe to all actions
	#  handler is a callback invoked when actions are received
}


interface CapUserPubSub {
# CapUserPubSub is the publish/subscribe capability available to Hub end-users.

	pubAction @0 (publisherID :Text, thingID :Text, actionID :Text, value :Data) -> ();
    # PubAction publishes an action request for a Thing.
	# Authorization will only allow actions to be published for things that are in the same group as the user
	# and for which the user has the operator or manager role.
	#  publisherID is the ID of the device or service that is publishing the thing
	#  thingID is the ID of the Thing whose action is being requested
	#  actionID is the ID as defined in the Thing's TD
	#  value is the JSON encoded value of the action

	subEvent @1 (publisherID :Text, thingID :Text, eventID :Text, handler :CapSubscriptionHandler) -> ();
	# SubEvent subscribes to events from a thing
	#  publisherID is the ID of the device or service that is publishing the thing event.
	#  thingID is the ID of the Thing whose event is published.
	#  eventID of the event. Use "" to subscribe to all events of the things.

	subTDs @2 (handler :CapSubscriptionHandler) -> ();
	# SubTDs subscribes to eligible TD events
	#  handler is a callback invoked when a TD is received

}



interface CapSubscriptionHandler {
# SubscriptionHandler is the callback interface for subscriptions

   handleValue @0 (value :Thing.ThingValue) -> ();
}
