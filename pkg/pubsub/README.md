# pubsub

The HiveOT Pub-Sub service enables services to publish and subscribe to event and action messages. This is a high-level capability for communicating with Thing Devices.

Protocol bindings can provide a bridge to 3rd party protocols such as MQTT and other message bus protocols. This is out of scope for this service.

## Status

This service is in the development stage. It is functional but still needs:
* queuing - IoT devices that reconnect receive new messages since last disconnect


## What problem does pub/sub solve?

The main problem to solve is that the publisher can send data to one or multiple receivers without needing to know who they are. As such an IoT device simply publishes events and subscribes to actions. This simplifies IoT devices as they don't need to concern themselves with authentication, authorization, web servers for user interface, and so on.

Similarly, publishers of actions to control IoT devices do not need to know how to connect to the IoT device. If fact, no-one is allowed to connect to IoT devices directly. Instead, publishers publish their messages on the message bus which routes it to the IoT device.


## Use cases

### IoT Device Publishes an Thing Value Event 

An IoT device uses MQTT to publish an event. The MQTT binding receives the event, validates the publisher, transforms the event if needed, and publishes the event on the internal pub/sub message bus. Services that subscribe to the event will receive the event in order of subscription.

If the event contains a TD then it is passed on to the directory service which will update its collection of TD documents.

The event is passed on to the history service which will update its history of the thing and update the property 'latest' value for user querying.

An automation service will subscribe to select event and runs the event through its automation rules. A web viewer will receive a notification through its https protocol adapter.

### Consumer Publishes a Thing Action 

A consumer wishes to turn the light on. The consumer uses a mobile phone app that connects and authenticates to the backend using https. The application posts a request to change the switch, which is received by the https protocol binding. 

The binding obtains a publication capability for the Thing from the pubsub service. This capability has a middleware hook for validating any requests. The binding publishes the action request and the capability runs it through the middleware chain. When validated the request is published. 

The IoT device is already connected and subscribed to action requests for its things. It receives the request via the subscribe capability and  changes the state of the switch. The state change result triggers publication of an event which confirms the action is complete.

### Protocol bindings such as MQTT 

MQTT is a popular message bus used with IoT. The MQTT binding provides a bridge between the internal message bus and the MQTT protocol. Messages from the MQTT bus are validated in the same way as messages from the HTTP gateway or IoT gateway.  

The MQTT binding lets consumers including web browsers work seamlessly with the Hub to receive events and publish actions.

This binding is a separate 'gateway' service for communicating with 3rd party clients and out of scope for this service.

### Authorization 

Authorization is a middleware task. 

The pub/sub capabilities accept middleware hooks to validate messages. Middleware can be used for authorization, rate limiting, logging, resiliency and other common tasks. 

Details TBD.

## Considerations

### MQTT and other message bus integration

MQTT is commonly used for pub/sub of IoT messages. A protocol adapter can use it to receive/publish messages and convert the protocol to that of the internal pub/sub. 

Services that use pub/sub remain independent of the protocol used to transport the message.

### Pub/sub vs Service, who depends on who? 

TL&DR: protocol binding use -> pubsub which pushes events/actions into -> storage

On the one hand,services such as the directory and history store only have a single responsibility, store the data. Collecting the data is not their responsibility.

On the other hand, the job of the pub/sub service is to route messages from the gateway that publishes it to subscribers. It should not have to know who the subscribers are. 

When new services are added, how are they receiving messages?

Therefore, who is responsible for subscribing to messages?

The choice is to put internal pubsub below the directory and history service. These services are responsible for receiving the events and actions. This allows for changing storage service implementations without affecting the Hub core. 


### Who is responsible to authorize publications and subscriptions?

Since the internal pubsub is application agnostic, it has no domain knowledge of the application. Its sole responsibility is to provide the capability to serve publish/subscribe requests.

When obtaining this capability, constraints can be included so that the publisher and subscribers are only allowed access to certain topics. Similarly, rate limiting can also be provided as a constraint.

The issuer of the capability is responsible for setting the constraints, eg provide the middleware to apply them.



### Who is responsible for rate limiting?

Rate limiting limits the number of messages that can be published per time period by a client.

Rate limiting can be added as a middleware task, injected in the publish capability.



## Design - capabilities based pub/sub

### Topic Format

The term topic comes from MQTT and defines the address of the message. A topic consists of multiple address parts: part1/part2/... 

In the Hub's internal pubsub, a topic is constructed as: things/{publisherID}/{thingID}/{msgType}/{name}. This will match the default schema used for the MQTT binding.
* where things is the prefix used to indicate the topic is that of a Thing event, action or td 
* where {publisherID} is the ThingID of the device that publishes the Thing.
* where {thingID} is the unique ID of the thing to subscribe to. ThingID's start with "urn:" as per WoT standard.
* where {msgType} is the type of message: "td", "event" or "action" as defined per vocabulary. 
* where {name} is the name of the event or action, or the devicetype of the thing TD. 

Note that this is for information only. The service API hides the topic construction.

### Subscribing

A Thing's device can only subscribe to actions of things it is the gateway for.
(eg: internal address things/{gatewayID}/+/action/+)

Services can subscribe to all events and actions of all things. For example, an automation service can use an action or event to trigger a rule. In order for a service to subscribe they must identify themselves as a Hub service. This depends on the method of connecting.

Users can only subscribe to events of things they are in the same group with. They must also be authenticated in order to subscribe. The authentication method depends on the connection protocol used and is handled by the protocol binding.

Users can subscribe to a single or multiple ThingIDs. However, they will only receive events from eligible Things, eg those they are in the same group with.

In the pubsub core implementation, a 'subscription', once approved, registers a callback that is invoked by pubsub when a message with the topic matches that of the subscription. 
Subscription supports multiple topics for the same callback.  

The pubsub core supports queuing of messages for a limited period of time. This is primarily intended for IoT devices with intermittent connectivity, for example devices that go to sleep. When a subscription is made by a device, messages that were published after its previous subscription disconnected, are forwarded. When subscribing the client can indicate the maximum age of messages that should be sent.    


### Publishing

IoT devices can publish events and TD documents. When a device publishes a message its thingID is used as the gatewayID. It can publish multiple things or just its own.

Services have capabilities of both IoT devices and Users, and can publish events and actions for any device.
When publishing events, or subscribing to actions, the service's thingID is used as the gatewayID.

Users can only publish actions. In addition they must be in the same group as the Thing whose action they are requesting. Last, their role in that group must allow it, eg 'operator' or 'manager'. Viewers, or IoT devices are not allowed to publish actions.  

### Implementation

This service is implemented in golang. 

The service API provides methods for obtaining device, service and user capabilities that are constrained to those usages. These capability enable publishing and subscribing to event and action messages. 

When publishing a message the capability passes the request through the middleware chain that handles tasks such as authorization, rate limiting, logging and others. The middlewware chain used is implemented separately from the pubsub service and can also be used in protocol binding services.  

When subscribing to an event or action, the core can send messages that were received in the interim since the last subscription was active. This is intended to allow devices with intermittent connections to receive action requests.  

For each incoming message the core service creates a goroutine to iterate the subscribers and pass the message to the helper for these subscribers.

The service has a capnp adapter that receives and forwards messages using the capnp protocol. Capnp supports callbacks which is used for subscription callbacks. The capnp transport layer depends on the Hub's configuration and can be Unix Domain Sockets, TCP sockets, or Named Pipes. 
