# Hub MQTT Gateway

The MQTT gateway is intended for clients that cannot use capnproto or require MQTT for integration. The main type of client is the web browser.

## Status

This service is in development


## Summary

Javascript clients do not have an easy way to establish a capnproto connection. This service provides a MQTT gateway to the hub, offering limited capabilities for use by javascript clients running in a web browser, or any other MQTT clients.

Features:
1. User based authentication
2. Subscribe to TD, events and actions
3. Publish TD, actions and events
4. Read directory [1] 
5. Read history [1]

As Mqtt is a pub/sub protocol, not a request/response protocol, responses to request for directory or history are send asynchroneously from the request. 

### Publish Action

Consumers can publish action requests from Things. The topic to use:
> PUBLISH: things/{publisherID}/{thingID}/action/{name}
 
... where name is the name of the action described in the Thing's TD.

For an action to be accepted, the client that publishes the event must have a role of operator in the group that both the user and the thing are a member of, as defined by authz. 


### Publish Event

Devices and services can publish events via MQTT in the same way actions are published.

> PUBLISH: things/{publisherID}/{thingID}/event/{name}


### Subscribe to Action

Publishers of devices and services can receive action requests on the topic:

> SUBSCRIBE: things/{publisherID}/{thingID}/action/{name}

Where:
* {publisherID} must be a valid publisher
* {thingID} is that of the Thing to activate
* {name} is that of the action to engage


### Subscribe to Event

To receive events from a Thing, consumers can subscribe to the event topic:
> SUBSCRIBE: things/{publisherID}/{thingID}/event/{name}

Where:
* {publisherID} must be a valid publisherm or '+' for any publisher
* {thingID} is that of a Thing or '+' for any Thing event from the publisher
* {name} is that of the event to subscribe to, or '+' for all events

At least one of the above parameters must be provided however. If all parameters are '+' the subscription is refused.


### Read Thing Directory

The request to read the Thing Directory is published on the following MQTT topic:

> things/directory/{thingID}/action/get
 
This requests the action 'get ThingID' from the directory service. ThingID can be a wildcard '+' which means all things in the directory.

The responses are published on things/directory/{thingID}/td **to the requesting client only**. The client has to subscribe to the topic and can use a wildcard '+' for {thingID}.

The directory service must be running.

### Read Thing Latest Values

To read the most recent values of a Thing, the client posts a request on the following MQTT topic:

> things/history/{thingID}/action/latest[/{name}]

Where [/{name}] is the optional event name whose value to get. By default, a properties event will all event values will be returned.

The response is published on things/{historyID}/{thingID}/event[/{name}]. If no name was given in the request then a properties event is returned containing a KV map with the most recent property values.

The history service must be running.


### Read Thing Value History

To read the history of a Thing event, the client posts a request on the following MQTT topic:

> things/history/{thingID}/action/history/{name}

Where {name} is the event name whose history to get.

The response is published on things/{historyID}/{thingID}/event/{name}, containing a map of {timestamp:value,...}. The default range is the last 24 hours.

To request a different time range, include a JSON obtain in the payload containing the start time and end time:

```json
{
  "start": "YYYY-MM-DDTHH:MM",
  "end": "YYYY-MM-DDTHH:MM",
  "limit": 1000
}
```

Where "start" is the ISO8601 start time for the time range, "end" is the ISO8601 end time of the time range, and limit the number of values to include. If omitted, a default of 1000 is assumed. 

Paging:
If the number of results equals limit, then repeat the request with the start time of the last received event. This will return the next batch of event values. This can be repeated until the number of results is less than limit or no results are returned.  

The history service must be running.


## Clients

The service implementation provides a golang and js native client that implements the directory and history APIs. The golang client is intended for testing while the JS client for use in browser or nodejs applications.


## Design Overview

This service is built around the [mochi embedded MQTT broker](https://github.com/mochi-co/mqtt)

The MQTT library is used as a gateway to the internal services. Publications are intercepted by this service and forwarded to the appropriate service. 

Subscriptions are constrained to the topics for things and services the user has access to. MQTT publication onto those topics by other clients are blocked. 

Effectively the MQTT broker is limited to client-gateway interaction and is constrained by the capabilities the gateway has afforded the client based on its credentials.

> MQTT client -> MQTT broker -> Mqtt session -> gateway session -> service capability.

Note that these constraints and redirects are transparent to the client. Clients can interact with the MQTT broker as normally would be the case.

**Latency Penalty:** 
The mqtt to gateway forwarding will have a latency penalty due to the extra hops from mqtt session (this service) to the gateway. If this service and gateway live on the same machine, this extra hop adds a latency in the order of 100usec - 1msec, depending on the machine resources. (this is a rough estimate based on other performance tests)


### Authentication

When a client connects to the MQTT broker, this service obtains a gateway capnp client instance for the duration of the connection. If the MQTT connection is closed for any reason, the capnp client is released. 

The client login credentials passed to MQTT are used to login to the gateway. If the gateway login fails, the MQTT session is closed.  

Instead of asking the user for a login password it is possible to use a saved refresh token for the mqtt connection, which the service uses to refresh the login to the gateway. A new refresh token is returned which is passed to the MQTT broker and returned to the client. If the token has expired the MQTT connection is closed and gateway client is released.


Authentication or authorization of publications are handled by the gateway service. The gateway can refuse a request due to various reasons:
1. The client is not authenticated
2. The client has no authorization to obtain the associated capability.
3. The client has no authorization to access the Thing.

If a request is refused, the publication is ignored when using MQTT 3 and an error is returned when using MQTT 5.

### Publish

When the client publishes onto a topic, this is intercepted by the service. The publication is never passed on to any other subscribers.

This service determines which capability is needed to handle the publication, for example the read directory capability, read history capability or publish capability from the pubsub service. 

On the first request, the needed capability is obtained from the gateway. On successive requests it is re-used. When the connection is closed, all capabilities are released.

### Subscribe

When the client subscribes to a topic, this is intercepted by the service. 

This service determines which capability is needed to handle the subscription, for example the read directory capability, read history capability or subscribe capability from the pubsub service.

On the first request, the needed capability is obtained from the gateway. On successive requests it is re-used. When the connection is closed, all capabilities are released.

