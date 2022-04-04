# MQTT protocol binding

This document specifies the MQTT protocol binding for the WoST Hub.

Bindings are based on the [W3C WoT bindings specification](https://w3c.github.io/wot-binding-templates/#creating-a-new-protocol-binding).

Documentation of a binding should contain:
* URI schema
* Mapping to WoT operations, e.g. "readproperty", "writeproperty", "invokeaction", ...
* Document that specifies the protocol.



## URI Schema

The MQTT protocol is identified with the 'mqtt://' URI schema.


## WoT Operations Mapping

Based on the draft from https://w3c.github.io/wot-binding-templates/bindings/protocols/mqtt/index.html.

| property operations     | binding                                 | topic                             |
|-------------------------|-----------------------------------------|-----------------------------------|
| readproperty            | mqv:controlPacketValue": "PUBLISH"      | things/{thingID}/read/properties  |
| readmultipleproperties  | mqv:controlPacketValue": "PUBLISH"      | things/{thingID}/read/properties  |   
| readallproperties       | mqv:controlPacketValue": "PUBLISH"      | things/{thingID}/read/properties  |
| writeproperty           | "mqv:controlPacketValue": "PUBLISH"     | things/{thingID}/write/properties |
| writemultipleproperties | "mqv:controlPacketValue": "PUBLISH"     | things/{thingID}/write/properties  |
| writeallproperties      | not supported                           | things/{thingID}/write/properties  |
| observeproperty         | "mqv:controlPacketValue": "SUBSCRIBE"   | things/{thingID}/event/properties|
| observeallproperties    | "mqv:controlPacketValue": "SUBSCRIBE"   |things/{thingID}/event/properties|
| unobserveproperty       | "mqv:controlPacketValue": "UNSUBSCRIBE" |things/{thingID}/event/properties|
| unobserveallproperties  | "mqv:controlPacketValue": "UNSUBSCRIBE" |things/{thingID}/event/properties|


## Property Operations

Below the WoST MQTT binding for standard operations. 

### readproperty (tentative)
The operation readproperty is used by a consumer to request to read a property of a Thing.

> This section is for consideration and not currently implemented

Form:
>```json
>{
>  "op": "readproperty",
>  "href": "mqtts://{broker}/things/{thingID}/read/properties",
>  "mqv:controlPacketValue": "PUBLISH",
>  "contentType": "application/json"
>}
>```
>Payload: JSON encoded array with 1 property name
> {
>    ["propertyName"]
> }

The Exposed Thing responds with a 'properties change event' as defined in the TD properties event:
> 
> Topic: **things/{thingID}/event/properties**
>
> Payload: map with name-value pair of the requested property.
>```json
>{
>  "propertyName": "value"
>}
>```

### readmultipleproperties (tentative)
The operation readmultipleproperties is used by a consumer to request to read select properties of a Thing.

> This section is for consideration and not currently implemented

>Form:
>```json
>{
>  "op": "readmultipleproperties",
>  "href": "mqtts://{broker}/things/{thingID}/read/properties",
>  "mqv:controlPacketValue": "PUBLISH",
>  "contentType": "application/json"
>}
>```
>Payload:
>* JSON encoded array of property names

The Exposed Thing responds with a 'properties change event' as defined in the TD properties event:
> Topic: **things/{thingID}/event/properties**
> 
> Payload: map with name-value pairs of the given properties.
>```json
>{
>  "property1Name": "value",
>  "property2Name": "value"
>}
>```

 
### readallproperties (tentative)
The operation readallproperties is used by a consumer to request to read all properties of a Thing.

> This section is for consideration and not currently implemented

>Form:
>```json
>{
>  "op": "readallproperties",
>  "href": "mqtts://{broker}/things/{thingID}/read/properties",
>  "mqv:controlPacketValue": "PUBLISH",
>  "contentType": "application/json"
>}
>```
>Payload: none. The lack of payload instructs to respond with all properties. 

The Exposed Thing responds with a 'properties change event' as defined in the TD properties event, containing the value of all properties:
>Topic: **things/{thingID}/event/properties**
>
>Payload:
> map with name-value pairs of all properties of a thing.
>```json
>{
>  "property1Name": "value",
>  "property2Name": "value"
>}
>```

### writeproperty, writemultipleproperties

> This section is for consideration and not currently implemented

The operations writeproperty(...ies) are used by a consumer to publish a request to modify one or more more properties on an ExposedThing. 

>Form:
>```json
>{
>  "op": ["writeproperty","writemultipleproperties"],
>  "href": "mqtts://{broker}/things/{thingID}/write/properties",
>  "mqv:controlPacketValue": "PUBLISH",
>  "contentType": "application/json"
>}
>```
>
>Payload: JSON encoded map with property name-value pairs, where value must match the description in the property affordance schema.
>
>```json
>{
>  "property1Name": "value",
>  "property2Name": "value"
>}
>```

Response:
When the property change is accepted, a property value change event will be sent.


### observeproperty, observeallproperties

To be notified of a property value changes, subscribe to property events. 

> Note: This protocol binding makes no distinction between subscribing to a single or all properties. A ConsumedThing implementation can map a specific property to a corresponding callback handler.

Form: 
>```json
>{
>  "op": ["observeproperty","observeallproperties"],
>  "href": "mqtts://{broker}/things/{thingID}/event/properties",
>  "mqv:controlPacketValue": "SUBSCRIBE",
>  "contentType": "application/json"
>}
>```
>* Where {thingID} is the thing ID or '+' to subscribe to properties changes from all Things.

Response: Observers will receive property change events from the thing with the given ID.

>Emit Property Change Form:
>```json
>{
>  "op": "emitpropertychange",
>  "href": "mqtts://{broker}/things/{thingID}/event/properties",
>  "mqv:controlPacketValue": "PUBLISH",
>  "contentType": "application/json"
>}
>```
>Payload: Property change events contain the JSON encoded map of property name-value pairs for each of the properties that have changed. The values are of type described by the property affordance in the TD and can be a string, number, integer, boolean, or an object with an additional map of name-value pairs.
> For example
>```json
>{
>  "property1Name" : "value1",
>  "property2Name" : "value2"
>}
>```


### unobserveproperty, unobserveallproperties

To end observing property changes, unsubscribe from the properties event. The protocol binding makes no distinction between subscribing to a single or all properties. A ConsumedThing implementation should only unsubscribe when there are no other property subscriptions in effect on the ConsumedThing.

>```json
>{
>  "op": ["unobserveproperty","unobserveallproperties"],
>  "href": "mqtts://{broker}/things/{thingID}/event/properties",
>  "mqv:controlPacketValue": "UNSUBSCRIBE",
>  "contentType": "application/json"
>}
>```
* Where {thingID} is the thing ID used when subscribing (observing).


## Event Operations

| event operations     | binding                                 |
|----------------------|-----------------------------------------|
| subscribeevent       | "mqv:controlPacketValue": "SUBSCRIBE"   |
| subscribeallevents   | "mqv:controlPacketValue": "SUBSCRIBE"   |
| unsubscribeevent     | "mqv:controlPacketValue": "UNSUBSCRIBE" |
| unsubscribeallevents | "mqv:controlPacketValue": "UNSUBSCRIBE" |

### subscribeevent

Subscribe to be notified of an event from an exposed thing. 

>```json
>{
>  "op": "subscribeevent",
>  "href": "mqtts://{broker}/things/{thingID}/event/{eventName}",
>  "mqv:controlPacketValue": "SUBSCRIBE",
>  "contentType": "application/json"
>}
>```
* Where {thingID} is the thing ID or '+' to subscribe to events from all Things.
* Where {eventName} is the event to subscribe to or '+' to subscribe to all events.

Response: Events of the Thing with the given ID and matching event name:

>Form used by Exposed Thing to send the event:
>```json
>{
>  "op": "emitevent",
>  "href": "mqtts://{broker}/things/{thingID}/event/{eventName}",
>  "mqv:controlPacketValue": "PUBLISH",
>  "contentType": "application/json"
>}
>```
>Payload: Events contain the JSON encoded value described by the EventAffordance in the TD.
> value can be a string, number, integer, boolean, or an object with a map of name-value pairs. 
> For example a map with values:
>```json
>{
>  "eventProperty1" : "value1",
>  "eventProperty2" : "value2"
>}
>```


### unsubscribeevent

Form:
>```json
>{
>  "op": "unsubscribeevent",
>  "href": "mqtts://{broker}/things/{thingID}/event/{eventName}",
>  "mqv:controlPacketValue": "UNSUBSCRIBE",
>  "contentType": "application/json"
>}
>```
* Where {thingID} is the thing ID used when subscribing
* Where {eventName} is the event used when subscribing.

Payload: no payload

## Action Operations

| Operation       | Binding                             |
|-----------------|-------------------------------------|
| invokeaction    | "mqv:controlPacketValue": "PUBLISH" |
| cancelaction    | not supported                       |
| queryaction     | not supported                       |
| queryallactions | not supported                       |

### invokeaction

Request an action to take place on a Thing.

Form:
>```json
>{
>  "op": "invokeaction",
>  "href": "mqtts://{broker}/things/{thingID}/action/{actionName}",
>  "mqv:controlPacketValue": "PUBLISH",
>  "contentType": "application/json"
>}
>```

* Where {thingID} is the thing ID of the Exposed Thing that handles the action.
* Where {actionName} is the action to invoke.

Payload: JSON encoded action input data as described by the ActionAffordances in the TD.
In WoST this typically includes a 'id' field.


>Response: ExposedThings can respond to an action by emitting an action status event with the output described by the ActionAffordance in the TD.
If used then the action status event must be defined with an EventAffordance using the actionName as the event name.
Form:
>```json
>{
>  "op": "emitevent",
>  "href": "mqtts://{broker}/things/{thingID}/event/{actionName}",
>  "mqv:controlPacketValue": "PUBLISH",
>  "contentType": "application/json"
>}
>```
>
The 'output' data schema of the action in the ActionAffordance describes the content of the action status event. If an 'id' field is included in the input then this will also be included in the output.

* Event for action status: started, completed, cancelled, failed
* Topic: things/{thingID}/event/{actionName}
* Payload: JSON encoded action status object with the following properties:
  * id: action id provided when emitting the action
  * name: the name of the action 
  * status: status of the action: started, completed, cancelled, failed
  * description: additional human description of the action status, such as error messages or other.

For example:
```json
{
  "id": "{actionID}",
  "name": "{actionName}",
  "status": "started",
  "description": "optional details of the status"
}
```


## Options

MQTT supports the following options. Note that these might not be supported in the current implementation. 

### retain - To Retain or Not To Retain

Many MQTT implementations support 'retained' messages. The last received message on a topic is stored and after subscribing to a topic, this last message for this topic is immediately received. It acts as a cache. 

'retain' MUST never be used on actions as this can lead to repeated actions.

The plus side of enabling retain for TDs, properties and possibly events, is that the most recent TD's and events will be received immediately on connecting to the message bus.

The downside is that this can lead to a lot of messages when consumers use a wildcard subscription. Not all clients handle the avalanche of messages gracefully. It can also cause significant and costly bandwidth consumption.

In WoST the recommendation for consumers is NOT to use retain unless there is a specific use-case to do so.  


### qos 

MQTT supports QOS of 0, 1 or 2. In WoST a default QOS of 1 is assumed (guaranteed delivery at least once).

A Qos of 0 can be used in case of high frequency updates of the same event, where intermittently dropped messages have little impact on the application.

For actions that are not idempotent a QOS of 2 (exactly once) should be used.

Example of a form schema in a TD:
```json
{
  "form": {
    "op": "writemultipleproperties",
    "contentType": "application/json",
    "topic": "things/{thingID}/properties",
    "mqv:controlPacketValue": "PUBLISH",
    "options": {
      "qos" : "1"
    }
  }
}
```

### dup

This flag is set by the protocol binding if a received message is marked as a duplicate by the MQTT broker.

It is currently ignored.
