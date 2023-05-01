package pubsub

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"

	"github.com/hiveot/hub/lib/thing"
)

const ServiceName = hubapi.PubsubServiceName

// A note on pubsub addressing:
// Any IoT device or service that publishes Thing events and listens for actions is a gateway
// for those Things. A gateway can host just one or multiple Things.
//
// The address format used in publishing and subscribing is:
//
//	things/{publisherID}/{thingID}/{messageType}[/{name}]
//
// Where:
//  things is the prefix for publishing Thing related data. The pubsub can be used for other internal purposes as well.
//  {publisherID} is the thingID of the publishing device. Eg urn:servicename where servicename is unique to the hub.
//  {messageType} is event, action or td
//  {name} is the event or action name, following the vocabulary standardized names
//         if messageType is a td then name is the device type
//
// Devices and services typically publish their own Thing TD on address:
//
//	things/{publisherID}/{publisherID}/{messageType}/{name}
//
// * Valid publisherIDs and thingIDs must start with "urn:" and contain only alphanum or ":_-." characters.
// * The character "+" is a wildcard characters for that part of the address.
// * Publishers listen for actions on the address {publisherID}/+/action/+
// * Publishers publish events on the address {publisherID}/{thingID}/event/{name}

// The IPubSubService interface provides a high level API to publish and subscribe actions and events
type IPubSubService interface {

	// CapDevicePubSub provides the capability to pub/sub Thing information as an IoT device.
	// This capability is only available to authenticated IoT devices.
	//
	// The deviceID is the thingID of the device and used as the publisherID for all TD's and
	// events published by this device.
	CapDevicePubSub(ctx context.Context, deviceID string) (IDevicePubSub, error)

	// CapServicePubSub provides the capability to pub/sub thing information as a hub service.
	// Hub services can publish their own information and receive events from any thing.
	// This capability is only available to authenticated Hub services.
	//
	// If the connection to the pubsub service fails then this capability becomes invalid and
	// must be obtained again of the connection is restored.
	//
	//  The serviceID is identifies the service publishing or subscribing
	CapServicePubSub(ctx context.Context, serviceID string) (IServicePubSub, error)

	// CapUserPubSub provides the capability for an end-user to publish or subscribe to messages.
	// This capability is only available to authenticated Hub users.
	//
	//  userID is the login ID of an authenticated user and is used as the publisherID
	CapUserPubSub(ctx context.Context, userID string) (IUserPubSub, error)
}

// IDevicePubSub available to an IoT device
type IDevicePubSub interface {
	// PubEvent publishes the given thing event. The payload is an event value as per TD.
	//
	// All published events will include the deviceID as the publisherID. In case the device
	// publishes its own events, its ID is both the publisherID and the thingID.
	//
	// The eventID is one of:
	// * the key in the events map of the TD when publishing an event
	// * hubapi.EventNameProperties when publishing a key-value map of changed properties
	// * hubapi.EventNameTD when publishing a TD document corresponding to the thingID
	//
	//  thingID of the Thing whose event is published
	//  eventID is one of the predefined events or the event's key in the TD event map
	//   TD documents are sent with the 'td' event name and property maps with the 'properties' event name.
	//  value is the serialized event value, or nil if the event has no value
	PubEvent(ctx context.Context, thingID, eventID string, value []byte) (err error)

	// Release the capability and end subscriptions
	Release()

	// SubAction registers a listener for actions to things managed by this device.
	//
	// The actionID is one of:
	// * the key in the actions map of the TD corresponding to the thingID, or
	// * hubapi.ActionNameConfiguration when receiving request to change writable properties.
	//   the data provided must contain a serialized key-value map with the properties and their new values.
	// * the key in the properties map of the TD when requesting a change to a single property
	//   the data provided is the serialized new value.
	//
	//  thingID is the ID of the Thing whose action to subscribe to, or "" for all things managed by this device.
	//  actionID is the action ID, the key in TD action map. Use "" to subscribe to all actions including properties
	//  handler will be invoked when an action is received for this device
	SubAction(ctx context.Context, thingID string, actionID string,
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

	// SubActions subscribes to actions aimed at things from any publisher.
	//
	// This is intended for services that track actions aimed at other devices or services,
	// Possibly a logging, monitoring or automation service.
	//
	//  publisherID is the ID of the publisher that is receiving the actions or "" for all publishers.
	//  thingID is the ID of the Thing whose action to subscribe to or "" for
	//   all things published by the publisher.
	//  actionID or "" to subscribe to all actions
	//  handler is a callback invoked when actions are received
	SubActions(ctx context.Context, publisherID, thingID string, actionID string,
		handler func(action *thing.ThingValue)) (err error)

	// SubEvents subscribes to events aimed at things from any publisher
	//
	// It is not recommended to subscribe to all events of all things of all publishers unless
	// absolutely needed for the task at hand, as this can generate a lot of calls.
	//
	// This is intended for services that track certain events from any device or service.
	// Possibly a logging, monitoring or automation service.
	SubEvents(ctx context.Context, publisherID, thingID string, eventID string,
		handler func(action *thing.ThingValue)) (err error)

	// Release the capability and end subscriptions
	Release()
}

// IUserPubSub defines the capability of an end-user to publish and subscribe messages
type IUserPubSub interface {
	// PubAction publishes an action request for a Thing.
	//
	// Authorization will only allow actions to be published for things that are in the same group as the user
	// and for which the user has the operator or manager role.
	// Managers and services are allowed to send the hubapi.ActionNameConfiguration action to
	// request a configuration change of the Thing.
	//
	//  publisherID is the ID of the device or service that is publishing the thing
	//  thingID is the ID of the Thing whose action is being requested
	//  actionID is the action ID as defined in the Thing's TD actions map
	//  data is the serialized value of the action
	// This returns an error if the action request could not be delivered.
	PubAction(ctx context.Context, publisherID, thingID, actionID string, data []byte) (err error)

	// SubEvent subscribes to events from a thing
	//
	// It is not allowed to subscribe to all events of all things of all publishers.
	// A thingID or an eventID must be provided or this will return an error.
	//
	//  publisherID is the ID of the device or service that is publishing the thing event.
	//  thingID is the ID of the Thing whose event is published.
	//  eventID of the event. Use "" to subscribe to all events of a thing.
	SubEvent(ctx context.Context, publisherID, thingID, eventID string,
		handler func(event *thing.ThingValue)) error

	// Release the capability and end subscriptions
	Release()
}
