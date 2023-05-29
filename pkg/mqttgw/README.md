# Hub MQTT Gateway

The MQTT gateway is intended for clients that cannot use capnproto or require MQTT for integration. The main type of client is the web browser.

## Status

This service is functional.

Known issues:
* certificate based authentication is in development
* JWT refresh token support
* authorization is in development
* directory paging is not yet supported
* The history response of a thing repeats publisherID, thingID and event name for every value, which is very inefficient. 
* The directory response encodes the TD then encodes the response in json, which is also inefficient. 

## Summary

Not all clients have an easy way to establish a Capnproto connection with the Hub. For example, there is no easy to use javascript serializer and RPC at the time of writing. To work arond this, this service provides a MQTT gateway to the hub that can be used over TCP and websocket connections. It offers the ability to users and devices for accessing some of the Hub capabilities. It is also useful to facilitate integration with MQTT capable systems.

Features:
1. Authentication using password, token, or peer certificates
2. Authorization using the authz service
3. Users can subscribe to events and publish actions
4. Devices can publish events and subscribe to actions
5. Users can read directory
6. Users can read history
7. A golang client for ease of use

As Mqtt is a pub/sub protocol, not a request/response protocol, responses to request for directory or history are send asynchroneously from the request. 

### Publish Thing Action Requests

Consumers can request actions of Things. The mqtt topic to use:
> PUBLISH: things/{publisherID}/{thingID}/action/{name}
 
... where {name} is the name of the action described in the Thing's TD.

For an action to be accepted, the client that publishes the event must have a role of operator in the group that both the user and the thing are a member of, as defined by authz. 


### Publish Thing Event

Devices and services can publish events via MQTT in the same way actions are published.

> PUBLISH: things/{publisherID}/{thingID}/event/{name}


### Subscribe to Thing Actions

Publishers of devices and services can receive action requests by subscribing to the topic:

> SUBSCRIBE: things/{publisherID}/{thingID}/action/{name}

Where:
* {publisherID} must be the authenticated publisher of the Thing
* {thingID} is that of the Thing to activate
* {name} is that of the action to engage


### Subscribe to Thing Events

To receive events from a Thing, consumers can subscribe to the event topic:
> SUBSCRIBE: things/{publisherID}/{thingID}/event/{name}

Where:
* {publisherID} must be a valid publisherm or '+' for any publisher
* {thingID} is that of a Thing or '+' for any Thing event from the publisher
* {name} is that of the event to subscribe to, or '+' for all events

At least one of the above parameters should be provided however. If all parameters are '+' the subscription can be refused.


### Read Thing Directory

The request to read the Thing Directory is published on the following MQTT topic:

> services/directory/action/directory

The payload must contain a JSON document with filter parameters:
```json
{
  "publisherID": "publisherID", // optional filter on publisher of the Thing
  "offset": 0,                  // starting offset for paging
  "limit": 1000,                // maximum number of results to return for paging
}
```

This requests the action 'read directory' from the directory service.

The responses are published on services/directory/event/directory **to the requesting client only**. The client has to subscribe to this topic.

The payload contains a list of ThingValue objects as returned by the directory service:
```json
{
  "tds": [
      thing.ThingValue...
  ],
  "itemsRemaining": false
}
```

...where a thing.ThingValue contains: 
```json
{
  "id": "td",
  "publisherID": "publisher thingID",
  "thingID": "thingID",
  "data": "{json encoded TD}",
  "created": "2023-05-22T07:00:00-0700"
}
```

The directory service must be running.

### Read Thing Value History

To request the history of a Thing event, the client posts a read action on the following MQTT topic:

> services/history/action/history

Where:
* history addresses the default history service.
* action indicates the message is an action request 
* read indications the request is to read history

The payload must contain a JSON document with filter parameters:
```json
{
  "publisherID": "publisherID",            // required publisher of the Thing
  "thingID": "thingID",                    // required thing whose history to get
  "name": "eventname",                     // required name of event whose history to get
  "startTime": "YYYY-MM-DDTHH:MM:SS.TZ",   // optional ISO8601, default 24 hours ago
  "duration": {seconds},                   // optional seconds. default is 24*3600
  "limit": 1000                            // optional max results to include, default is 1000
}
```

The response is published on services/history/event/history. The payload is a JSON object, containing a time ordered list of ThingValue objects.


```json
{
  "itemsRemaining": "false",
  "name": "eventname",
  "publisherID": "publisherID",
  "thingID": "thingID",
  "values": [
    {
      "id": "eventname",
      "publisherID": "publisherID",
      "thingID": "thingID",
      "created": "timestamp1",
      "data": "value1",
    },
    ...
  ]
}
```

Paging:
If itemsRemaining is true then repeat the request with the start time of the last received event. This will return the next batch of event values. This can be repeated until 'itemsRemaining' is false.  

The history service must be running.


### Read Thing Latest Values

To request the most recent property or event values of a Thing, the client posts a request on the following MQTT topic:

> services/history/action/latest

The payload is a JSON document with filter parameters:
```json
{
  "publisherID": "publisherID",  // filter by publisher
  "thingID": "thingID",          // filter by thing ID
  "names": []                    // list of event names to get. Default is all
}
```

The response is a list of value objects published on:
> services/history/event/latest

containing the payload:
```json
{
  "publisherID" : "publisherID",
  "thingID": "thingID",
  "values": [   
    {
      "name": "eventName",
      "created": "2023-05-06T11:00:53-07:00",
      "data": "...",    
    },
  ...
  ]}
```
 
The history service must be running.


## Clients

The service implementation provides a golang and js native client that implements the directory and history APIs. The golang client is intended for testing while the JS client for use in browser or nodejs applications.


## Design Overview

This service is built around the [mochi embedded MQTT broker](https://github.com/mochi-co/mqtt)

The MQTT library is used as a gateway to the internal services. Publications are intercepted by this service and forwarded to the appropriate service. 

Subscriptions are constrained to the topics for things and services the user has access to. MQTT publication onto those topics by other clients are blocked. 

Effectively the MQTT broker is limited to client-gateway interaction and is constrained by the capabilities the gateway has afforded the client based on its credentials.

> MQTT client -> Mochi-co mqtt broker -> mqttgw session -> gateway session -> service capability.

Note that these constraints and redirects are transparent to the client. Clients can interact with the MQTT broker as normally would be the case.

**Latency Penalty:** 
The mqtt to gateway forwarding will have a latency penalty due to the extra hops from mqtt session (this service) to the gateway. If this service and gateway live on the same machine, this extra hop adds a latency in the order of 100usec, depending on the machine resources. (this is a rough estimate based on other performance tests)


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

This service determines which capability is needed to handle the publication, for example the read directory capability, read history capability or publish capability from the pubsub service. Pubsub publications are passed on to the pubsub service. Directory and History publications will invoke these services.    

On the first request, the needed capability is obtained from the gateway. On successive requests it is re-used. When the connection is closed, all capabilities are released.

### Subscribe

When the client subscribes to a topic, this is intercepted by the service. Pubsub subscriptions are passed on to the pubsub service. All mqtt subscriptions are passed on to the broker. 

When a pubsub subscription is receiving a publication on the Hub message bus, it is passed on to the remote mqtt client via Mochi-co's direct message feature.  

On the first request, the needed capability is obtained from the gateway. On successive requests it is re-used. When the connection is closed, all capabilities are released.

