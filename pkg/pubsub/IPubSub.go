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
	CapDevicePubSub(ctx context.Context, deviceID string) IDevicePubSub

	// CapServicePubSub provides the capability to pub/sub thing information as a hub service.
	// Hub services can publish their own information and receive events from any thing.
	CapServicePubSub(ctx context.Context, serviceID string) IServicePubSub

	// CapUserPubSub provides the capability for an end-user to publish or subscribe to messages.
	// The caller must authenticate the user and provide appropriate configuration.
	//  userID is the login ID of an authenticated user
	CapUserPubSub(ctx context.Context, userID string) (pub IUserPubSub)

	// Stop the service and free its resources
	Stop(ctx context.Context) error
}

// IUserPubSub defines the capability of an end-user to publish and subscribe messages
type IUserPubSub interface {
	// PubAction publishes an action request for a Thing.
	// Authorization will only allow actions to be published for things that are in the same group as the user
	// and for which the user has the operator or manager role.
	//  action to send
	PubAction(ctx context.Context, action *thing.ThingValue) (err error)

	// SubEvent creates a topic for receiving an event
	//  publisherID of the event. Use "" to subscribe to all publishers
	//  thingID of the publisher event. Use "" to subscribe to events from all Things
	//  eventName of the event. Use "" to subscribe to all events of publisher things
	SubEvent(ctx context.Context, gatewayID, thingID, eventName string,
		handler func(event *thing.ThingValue)) error

	// SubTDs subscribes to all TD events.
	//  gatewayID is optional. Use to limit subscriptions to TDs from a particular gateway.
	//  handler is a callback invoked when a TD is received from a thing's publisher
	SubTDs(ctx context.Context, gatewayID string, handler func(event *thing.ThingValue)) (err error)

	// Release the capability and end subscriptions
	Release()
}

// IDevicePubSub available to an IoT device
type IDevicePubSub interface {
	// PubEvent publishes the given thing event. The payload is an event value as per TD.
	PubEvent(ctx context.Context, thingEvent *thing.ThingValue) (err error)

	// PubProperties creates a topic and publishes properties of a thing.
	// The props is a map of property name-value pairs.
	PubProperties(ctx context.Context, thingID string, props map[string]string) (err error)

	// PubTD publishes the given thing TD. The payload is a serialized TD document.
	PubTD(ctx context.Context, thingID string, deviceType string, td []byte) (err error)

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
	//  gatewayID of the thing. Use "" to subscribe to all gateways
	//  thingID of the action target. Use "" to subscribe to all Things of the gateway
	//  actionName or "" to subscribe to all actions
	//  handler is a callback invoked when actions are received
	SubActions(ctx context.Context,
		gatewayID string,
		thingID string,
		actionName string,
		handler func(action *thing.ThingValue)) (err error)

	// Release the capability and end subscriptions
	Release()
}
