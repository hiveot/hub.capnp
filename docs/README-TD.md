# WoST use of the TD - under development

WoST intends to be compliant with the WoT Thing Description specification.
The latest known draft at the time of writing is [Mar 2022](https://www.w3.org/TR/wot-thing-description11/#thing)

Interpreting this specification is done as a best effort. In case where discrepancies are reported they will be corrected when possible as long as they do not conflict with the WoST core paradigm of 'Things do not run servers'.

The WoT specification is closer to a framework than an application. As such it doesn't dictate how an application should use it. This document describes how the WoST information model and behavior maps to the WoT TD.   


# WoST IoT Device Model


## I/O Devices, Gateways and Publishers are IoT 'Things'

Most IoT devices are pieces of hardware that have embedded software that manages its behavior. Virtual IoT devices are build with software only but are otherwise considered identical to hardware devices.

IoT devices often fulfill multiple roles: a part provides network access, a part provides access to inputs and outputs, a part reports its state, and a part that manages its configuration.

WoST makes the following distinction based on the primary role of the device:
* A gateway is a Thing that provides access to other independent Things. A Z-Wave controller USB-stick is a gateway that uses the Z-Wave protocol to connect to I/O devices. A gateway is independent of the Things it provides access to and can have its own inputs or outputs. Gateways are often used when integrating with non-WoST Things. 
* A publisher is a Thing that publishes other Thing information to the WoST Hub. Publishers have their own ID that is included as part of the Thing ID of all Things that it publishes. A publisher has by default  authorization to publish and subscribe to the things it is the publisher of. Publishers often are services that convert between the WoT/WoST standards and a native protocol.
* An I/O device is Thing whose primary role is to provide access to inputs and outputs and has its own attributes and configuration. In case of hybrid hardware where attributes and configuration are managed by a parent device then the inputs and outputs are also considered to be part of the parent device.
* A Hub bridge is a device that connects two Hubs and shares Thing information between them. 

## Thing Description Document (TD)

The Thing Description document is a [W3C WoT standard](https://www.w3.org/TR/wot-thing-description11/#thing) to describe Things. TDs that are published on the Hub MUST adhere to this standard and use the JSON representation format.  

The TD consists of a set of attributes and properties that WoST uses to describe WoST things and their capabilities.

## TD Attributes

TD Attributes that WoST uses are as follows. Attributes marked here as optional in WoT are recommended in WoST:

| name               | mandatory | description                                    |
|--------------------|-----------|------------------------------------------------|
| @context           | mandatory | "http://www.w3.org/ns/td"                      |
| @type              | optional  | Device type from the WoST vocabulary           |
| id                 | mandatory | Unique Thing ID following the WoST ID format   |
| title              | mandatory | Human readable description of the Thing        |
| modified           | optional  | ISO8601 date this document was last updated    |
| properties         | optional  | Attributes and Configuration objects           |
| version            | optional  | Thing version as a map of {'instance':version} |
| actions            | optional  | Actions objects supported by the Thing         |
| events             | optional  | Event objects as submitted by the Thing        |
| security           | mandatory | "nosec". Not applicable in WoST                |
| securityDefinition | mandatory | empty. Not applicable in WoST.                 | 

Note: that security is not applicable to TD of IoT devices as they cannot be connected to. When services that provide a REST API are published as Things, the security attribute must be used to describe the security schema of the API.  


## WOST Thing ID Format

All Things have an ID that is unique to the Hub. To support uniqueness, authorization, and sharing with other Hubs, a WoST ID is constructed as follows:

> urn:{zone}/{publisherID}/{deviceID}/{deviceType}

* urn: prefix defined by the WoT standard 
* {zone} is the zone where the thing originates. All local things are in the 'local' zone. A bridge will change the zone of shared Things to that of the bridge domain.
* {publisherID} is the deviceID of the IoT device that provides access to a Thing. It must be unique on the Hub it is publishing to. Publishers are WoST compliant and implement the WoT/WoST standard. IoT devices that publish their own Thing are also publishers. In that case the publisherID and deviceID are the same and must be unique on the Hub. 
* {deviceID} is the ID of the hardware or software Thing being accessed. It must be unique on the publisher that is publishing it. In case of legacy devices this is the unique ID of the hardware. In case of an IoT Device that publishes its own Thing it is the same as the publisherID.
* {deviceType} is the type of Thing as defined in the [WoST vocabulary](https://github.com/wostzone/hub/blob/main/lib/client/pkg/vocab/IoTVocabulary.go). See the constants with prefix DeviceType. It describes the primary role of the device and is intended for filtering subscriptions on the message bus. 

When integrating with 3rd party systems that use a URI as the ID then the ID can be used as-is. If the ID is not a URI then it must be used as the deviceID, while the publisherID is that of the service that provides the protocol binding. Using a 3rd party ID as-is can lead to reduced capabilities for bridging, queries in the directory and other services.

The Thing ID is created by the publisher of a Thing.

## Thing Attributes (WoT read-only Properties)

Thing Attributes describe the Thing and the state it is in. They can only be read. For example device type and version are attributes. In a WoT TD these are included in the 'Properties' section. The @type attribute is used to indicate the WoT Property as an attribute, eg: "wost:attr'. If the @type attribute is omitted 'attribute' is assumed. 

Each attribute consists of a properties that describe it:

These attributes are derived from the [interaction affordance](https://www.w3.org/TR/wot-thing-description11/#interactionaffordance) and [dataschema](https://www.w3.org/TR/wot-thing-description11/#dataschema) sections of the WoT TD specification. 

| Attribute   | WoT       | description                                                          |
|-------------|-----------|----------------------------------------------------------------------|
| name        | optional  | Name used to identify the attribute in the TD Properties object. (1) |
| @type       | optional  | Type of the attribute: eg "wost:attr" for attributes. (2)            |
| type        | optional  | data type: string, number, integer, boolean, object, array, or null  |
| title       | optional  | Human description of the attribute.                                  |
| description | optional  | In case a more elaborate description is needed for humans            |
| forms       | mandatory | Tbd. currently not used in WoST                                      | 
| value       | optional  | Value of the attribute.                                              |
| minimum     | optional  | Minimum range value for numbers                                      |
| maximum     | optional  | Maximum range value for numbers                                      |
| enum        | optional  | Restricted set of values                                             |
| unit        | optional  | unit of the value                                                    |
| readOnly    | optional  | true for properties that are attributes, false implies writable      |
| writeOnly   | optional  | true for properties that are writable, eg configuration              |
| default     | optional  | Default value to use if no value is provided                         |

Notes:
1. Attribute names are standardized as part of the vocabulary so consumers can understand their purpose. The name is used in the TD Document Property list and can be included in the attribute definition for readability.
2. The '@type' attribute is used to identify the type of property (different from the data type), as suggested by @sebastiankb in this discussion: https://github.com/w3c/wot-thing-description/issues/1079. Property types have the "wost:" prefix are defined in hubapi api/vocabulary.go, eg: "wost:input", "wost:output", "wost:configuration", "wost:state", and "wost:attr". Configuration is the only writable attribute, state reflects the internal device state and can change in runtime, and attr is a static descriptive attribute such as vendor, version, and such.  When omitted, wost:attr is assumed.
3. The PropertyAffordance description in the WoT specification doesn't seem to apply to WoST. It is therefore ignored. The observation mechanism is handled by the Hub pub/sub message bus. WoST uses the WoT events to notify of changes. 
4. Attribute values can change as a result of an action. For example, upgrading firmware will change the device version attribute value. 
5. Device state, input and output values can be distinguished from regular read-only attributes by using the @type "wost:input", "wost:output", "wost:state" values.


## Thing Configuration (WoT writable Properties)

Configurable properties are writable and as such have read-only set to false and write-only to true.
To change configuration a 'configure' action is published. TBD. 

## Events

Changes to the state of a Thing are published using Events. The TD describes the events that a Thing publishes in its events section and is serialized as JSON in the following format:

```json
{
  "events" : {
    "{eventName}" : {
      ...InteractionAffordance,
      "data": {dataSchema},
      "dataResponse: {EventResponseData}",
    }
  }
}
```
Where:

* eventName: The name the event is published as. WoST predefines common events for changes to properties.
* data: Defines the data schema of event messages. The content follows the [dataSchema](https://www.w3.org/TR/wot-thing-description11/#dataschema) format, similar to properties.
* dataResponse: Describes the data schema of a possible response to the event. EventResponses are currently not used in WoST.

The [TD EventAffordance](https://www.w3.org/TR/wot-thing-description11/#eventaffordance) also describes optional subscription and cancellation attributes. These are not used in WoST as subscription is not handled by a Thing bu* by:hD MQTT message bus. # Property Event.

### The "property" Event

In WoST, changes to properties values (attribute, configuration, outputs), are signalled using events. For this purpose WoST defines a 'property' event. The 'property' event is required to be implemented by all WoST compatible devices. It does not have to be included in the 'events' section of the TD as it is part of the overall WoST schema and vocabulary.

The "property" event has the following schema:

```json
{
  "events": {
    "property": {
      "data": {
        "title": "Map of property name and new value pairs",
        "type": "object"
      }
    }
  }
}
```

For example, when a temperature has changed to 21 degrees and humidity to 55%, the event looks like this.
```json
{
  "events":{
    "property": {
      "temperature" : "21",
      "humidity" : "55"
    }
  }
}
```


## Actions

All interaction with Things take place via Actions. Actions are primarily used to change the value of configuration properties and to control inputs.

Actions are defined in the Thing Description document through [action affordances](https://www.w3.org/TR/wot-thing-description/#actionaffordance) 

```json
{
  "actions": {
    "{actionName"}: {ActionAffordance}
  }
}
```
Where actionName is a unique identification of the action, and ActionAffordance describes the action details:
```json
{
   ...interactionAffordance,
   "input": {},
   "output": {},
   "safe": true|false,
   "idempotent": true|false,
}
```

For example, a simple switch might be defined as:
```json
{ 
  "actions": {
    "onoff": {
      "title": "Control the on or off status of a switch",
      "input": {
        "type": "boolean"
      }
    }
  }
}
```

The actual action message to turn the switch on then looks like this:
{
  "actions": {
    "onoff": true
  }
}


### The "property" Action

WoST defines a 'property' action that must be supported by all WoST Things. This action updates the value of a writable property. It does not have to be included in the 'action' section of the TD as it is part of the overall WoST schema and vocabulary. The following schema is used:

```json
{
  "actions": {
    "property": {
      "title": "Update the value of writable properties",
      "input": {
        "type": "object",
        "idempotent": "true"
      }
    }
  }
}
```

For example, when the Thing configuration property called "name" changes, the action looks like this.
```json
{
  "action":{
    "property": {
      "name" : "Brand new name",
    }
  }
}
```

## Links

The spec describes a [link](https://www.w3.org/TR/wot-thing-description11/#link) as "A link can be viewed as a statement of the form "link context has a relation type resource at link target", where the optional target attributes may further describe the resource"

In WoST a link can be used as long as it is not served by the IoT device, as this would conflict with the paradigm that "Things are not servers". 


## Forms

The WoT specification for a [Form](https://www.w3.org/TR/wot-thing-description11/#form) says: A form can be viewed as a statement of "To perform an operation type operation on form context, make a request method request to submission target" where the optional form fields may further describe the required request. 

The provided example shows an HTTP POST to write a property. 

In WoST an important constraint is that form operations that define HTTP operations on the Thing device are prohibited. Things don't run servers and can therefore not respond to HTTP commands.
What is possible however is a form describing a MQTT topic on the Hub message bus, on which the device publishes or subscribes. Also allowed are forms that describe additional help information that is provided via an external http server.

This is tbd as the whole idea of forms and how they should be used is kinda murky.

### SecuritySchema 'scheme' (1)

In WoST all authentication and authorization is handled by the Hub. Therefore, the security scheme section only applies to Hub services and does not apply to WoST Things.


# REST APIs

WoST compliant Things do not implement servers. All interaction takes place via WoST Hub services and message bus. Therefore this section only applies to Hub services that provide a web API. For example, the Directory Service and Provisioning Service provide web API's.

Hub services that implement a REST API follows the approach as described in Mozilla's Web Thing REST API](https://iot.mozilla.org/wot/#web-thing-rest-api). 

```http
GET https://address:port/things/{thingID}[/...]
```

Note 1: the Mozilla specification often assumes or suggests that Things are directly accessed, which is not allowed in WoST. Therefore the implementation of this API in WoST MUST follow the following rules:
1. The Thing address is that of the hub it is connected to.
2. The full thing ID must be included in the API. The examples says 'lamp' where a Thing ID is a URN: "urn:zone:publisher:lamp" for example.
3. Things cannot include action href fields in their TD, only MQTT urls. 

Note 2: The Mozilla API does not support queries.

Note 3: Similar to the REST API, WebSocket based access follows Mozilla's specification. In addition, the websocket API allows listening for events. At the time of writing no websocket API's are in use so this section is still subject to change.

