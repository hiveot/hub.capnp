package pubsub

import (
	"context"

	"github.com/hiveot/hub.go/pkg/thing"
)

const ServiceName = "pubsub"

// A note on pubsub addressing:
// Any IoT device or service that publishes Thing events and listens for actions is a gateway
// for those Things. A gateway can host just one or multiple Things.
//
// The address format used in publishing and subscribing is:
//
//	things/{gatewayID}/{thingID}/{messageType}[/{name}]
//
// Where:
//  things is the prefix for publishing Thing related data. The pubsub can be used for other internal purposes as well.
//  {gatewayID} is the thingID of the gateway itself. Eg urn:servicename where servicename is unique to the hub.
//  {messageType} is event, action or td
//  {name} is the event or action name, following the vocabulary standardized names
//         if messageType is a td then name is the device type
//
// Gateways typically publish their own thing as well on address:
//
//	things/{gatewayID}/{gatewayID}/{messageType}/{name}
//
// * Valid GatewayIDs and ThingIDs must start with "urn:" and contain only alphanum or ":_-." characters.
// * The character "+" is a wildcard characters for that part of the address.
// * Gateways listen for actions on the gateway address {gatewayID}/+/action/+
// * Gateways publish events on the gateway address {gatewayID}/{thingID}/event/{name}

// Constants for constructing a gateway address
const (
	ThingsPrefix      = "things"
	MessageTypeAction = "action"
	MessageTypeEvent  = "event"
	MessageTypeTD     = "td"
)

// The IPubSubService interface provides a high level API to publish and subscribe actions and events
type IPubSubService interface {

	// CapDevicePubSub provides the capability to pub/sub thing information as an IoT device.
	// The issuer must only provide this capability after verifying the device ID.
	// The deviceID is the thingID of the device requesting the capability.
	CapDevicePubSub(ctx context.Context, deviceID string) (IDevicePubSub, error)

	// CapServicePubSub provides the capability to pub/sub thing information as a hub service.
	// Hub services can publish their own information and receive events from any thing.
	// The serviceID is the thingID of the service requesting the capability.
	CapServicePubSub(ctx context.Context, serviceID string) (IServicePubSub, error)

	// CapUserPubSub provides the capability for an end-user to publish or subscribe to messages.
	// The caller must authenticate the user and provide appropriate configuration.
	//  userID is the login ID of an authenticated user
	CapUserPubSub(ctx context.Context, userID string) (IUserPubSub, error)
}

// IDevicePubSub available to an IoT device
type IDevicePubSub interface {
	// PubEvent publishes the given thing event. The payload is an event value as per TD.
	// This will combine the thingID with the device's thingID to publish it under the thing address
	//  thingID of the Thing whose event is published
	//  name is the event name
	//  value is the serialized event value, or nil if the event has no value
	PubEvent(ctx context.Context, thingID, name string, value []byte) (err error)

	// PubProperties creates a topic and publishes properties of a thing.
	// This will combine the thingID with the device's thingID to publish it under the thing address
	//  thingID of the Thing whose event is published (not the thing address)
	//  The props is a map of property name-value pairs.
	PubProperties(ctx context.Context, thingID string, props map[string][]byte) (err error)

	// PubTD publishes the given thing TD. The payload is a serialized TD document.
	// This will combine the thingID with the device's thingID to publish it under the thing address
	//  thingID of the Thing whose event is published (not the thing address)
	PubTD(ctx context.Context, thingID string, deviceType string, tdDoc []byte) (err error)

	// Release the capability and end subscriptions
	Release()

	// SubAction creates a topic and registers a listener for actions to things with this gateway.
	// This supports receiving queued messages for this gateway since it last disconnected.
	//  thingID is the thing to subscribe for, or "" to subscribe to all things of this gateway
	//  name is the action name, or "" to subscribe to all actions
	//  handler will be invoked when an action is received for this device
	SubAction(ctx context.Context, thingID string, name string,
		handler func(action *thing.ThingValue)) (err error)
}

// IServicePubSub is the publish/subscribe capability available to Hub services.
// Hub services have IoT device capabilities and consumer capabilities as publishers of their own service and can
// subscribe similar to consumers. In addition to all events, actions and TDs.
type IServicePubSub interface {
	// IDevicePubSub allows services to publish as a Thing gateway
	IDevicePubSub

	// IUserPubSub allows services to consume other things
	IUserPubSub

	// SubActions subscribes to actions aimed at things.
	// Services can subscribe to other actions for logging, automation and other use-cases.
	// For subscribing to service directed actions, use SubAction.
	//
	//  thingAddr of the action target. Use "" to subscribe to all Things
	//  actionName or "" to subscribe to all actions
	//  handler is a callback invoked when actions are received
	SubActions(ctx context.Context,
		thingAddr string,
		actionName string,
		handler func(action *thing.ThingValue)) (err error)

	// Release the capability and end subscriptions
	Release()
}

// IUserPubSub defines the capability of an end-user to publish and subscribe messages
type IUserPubSub interface {
	// PubAction publishes an action request for a Thing.
	// Authorization will only allow actions to be published for things that are in the same group as the user
	// and for which the user has the operator or manager role.
	//  thingAddr is the address of the Thing whose action is being requested
	//  name is the action name as defined in the Thing's TD
	//  value is the JSON encoded value of the action
	//PubAction(ctx context.Context, action *thing.ThingValue) (err error)
	PubAction(ctx context.Context, thingAddr, actionName string, value []byte) (err error)

	// SubEvent subscribes to events from a thing
	//  thingAddr to subscribe to.
	//  eventName of the event. Use "" to subscribe to all events of the things.
	SubEvent(ctx context.Context, thingAddr, eventName string,
		handler func(event *thing.ThingValue)) error

	// SubTDs subscribes to eligible TD events
	//  handler is a callback invoked when a TD is received
	SubTDs(ctx context.Context, handler func(event *thing.ThingValue)) (err error)

	// Release the capability and end subscriptions
	Release()
}
