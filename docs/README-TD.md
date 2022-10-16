# HiveOT use of the TD - under development

HiveOT intends to be compliant with the WoT Thing Description specification.
The latest known draft at the time of writing is [Mar 2022](https://www.w3.org/TR/wot-thing-description11/#thing)

Interpreting this specification is done as a best effort. In case where discrepancies are reported they will be corrected when possible as long as they do not conflict with the HiveOT core paradigm of 'Things do not run servers'.

The WoT specification is closer to a framework than an application. As such it doesn't dictate how an application should use it. This document describes how the HiveOT information model and behavior maps to the WoT TD.

# HiveOT IoT Device Model

## I/O Devices, Gateways and Publishers are IoT 'Things'

Most IoT devices are pieces of hardware that have embedded software that manages its behavior. Virtual IoT devices are build with software only but are otherwise considered identical to hardware devices.

IoT devices often fulfill multiple roles: a part provides network access, a part provides access to inputs and outputs, a part reports its state, and a part that manages its configuration.

HiveOT makes the following distinction based on the primary role of the device. These are identified by their device type:

* A gateway is a Thing that provides access to other independent Things. A Z-Wave controller USB-stick is a gateway that uses the Z-Wave protocol to connect to I/O devices. A gateway is independent of the Things it provides access to and can have its own inputs or outputs. Gateways are often used when integrating with non-HiveOT Things.
* A publisher is a Thing that publishes Thing information to the HiveOT Hub. Publishers have a publishing ID that is included as part of the Thing ID of all Things that it publishes. A publisher has authorization to publish and subscribe to the things it is the publisher of.
* An I/O device is Thing whose primary role is to provide access to inputs and outputs and has its own attributes and configuration. 
* A Hub bridge is a device that connects two Hubs and shares Thing information between them.

## Thing Description Document (TD)

The Thing Description document is a [W3C WoT standard](https://www.w3.org/TR/wot-thing-description11/#thing) to describe Things. TDs that are published on the Hub MUST adhere to this standard and use the JSON representation format.

The TD consists of a set of attributes and properties that HiveOT uses to describe HiveOT things and their capabilities.

## TD Attributes

TD Attributes that HiveOT uses are as follows. Attributes marked here as optional in WoT are recommended in HiveOT:

| name               | mandatory | description                                     |
|--------------------|-----------|-------------------------------------------------|
| @context           | mandatory | "http://www.w3.org/ns/td"                       |
| id                 | mandatory | Unique Thing ID following the HiveOT ID format  |
| title              | mandatory | Human readable description of the Thing         |
| modified           | optional  | ISO8601 date this document was last updated     |
| properties         | optional  | Attributes and Configuration objects            |
| version            | optional  | Thing version as a map of {'instance':version}  |
| actions            | optional  | Actions objects supported by the Thing          |
| events             | optional  | Event objects as submitted by the Thing         |
| security           | mandatory | Protocol dependent security (see note1)         |
| securityDefinition | mandatory | Protocol dependent security (see note1)         | 

note: Consumers do not connect directly to the IoT device and authentication & authorization is handled by the Hub services. As a result the security definitions in the TD depend on the method used to access the HiveOT Hub, not the IoT device.

* HiveOT compatible IoT devices can simply use the 'nosec' security type when creating their TD and use a NoSecurityScheme as securityDefinition.
* Consumers, which access devices via 'Consumed Things', only need to know how connect to the Hub service. No knowledge of the IoT device protocol is needed.

## HiveOT Thing ID Format

All Things have an ID that is unique to the Hub. To support uniqueness, authorization, and sharing
with other Hubs, a HiveOT ID is constructed as follows:

HiveOT compatible IoT devices publish their 'Thing' with the following ID:
> urn:{deviceID}/{deviceType}

Protocol binding or services that publish other 'Things', for example a ZWave protocol binding, include the publisher ID using the following format:
> urn:{publisherID}/{deviceID}/{deviceType}

A bridge that publishes 'Things' from other Hubs include a 'zone' of the originating bridge using the following format:
> urn:{zone}/{publisherID}/{deviceID}/{deviceType}

* urn: prefix defined by the WoT standard
* {zone} optional zone where the Thing originates. All local things are in the 'local' zone. A bridge that receives Things from another Hub sets the zone to that of the originating Hub's domain. The local zone can be omitted.
* {publisherID} is the deviceID of the IoT device that provides access to a Thing. It must be unique on the Hub it is publishing to. Publishers MUST be HiveOT compliant and implement the WoT/HiveOT standard and vocabulary. IoT devices that publish their own Thing can omit the publisher. In that case the publisherID and deviceID are the same and must be unique on the Hub.
* {deviceID} is the ID of the hardware or software Thing being accessed. It must be unique on the publisher that is publishing it. In case of protocol bindings this is the ID of the original protocol. In case of a HiveOT compatible IoT Device this can be a mac address or other locally unique feature of the device.
* {deviceType} is the type of Thing as defined in
  the [HiveOT vocabulary](https://github.com/hiveot/hub.go/blob/main/pkg/vocab/IoTVocabulary.go). See the constants with prefix DeviceType. It describes the primary role of the device.

When integrating with 3rd party systems that use a URI as the ID, the ID can be used as-is. If the ID is not a URI then it must be used as the deviceID, while the publisherID is that of the service that provides the protocol binding. Using a 3rd party ID as-is can lead to reduced capabilities for bridging, queries in the directory and other services.

The Thing ID is created by the publisher of a Thing.

## Thing Properties

Thing Properties describe the Thing and the state it is in. For example, device type and version are properties. Read-only properties are considered attributes while writable properties are configuration.

The WoT TD describes properties with the [PropertyAffordance](https://www.w3.org/TR/wot-thing-description11/#propertyaffordance). This is a sub-class of an [interaction affordance](https://www.w3.org/TR/wot-thing-description11/#interactionaffordance) and [dataschema](https://www.w3.org/TR/wot-thing-description11/#dataschema).

HiveOT uses the following attributes to describe properties.

| Attribute   | WoT       | description                                                               |
|-------------|-----------|---------------------------------------------------------------------------|
| name        | optional  | Name used to identify the property in the TD Properties object. (1)       |
| type        | optional  | data type: string, number, integer, boolean, object, array, or null       |
| title       | optional  | Human description of the property.                                        |
| description | optional  | In case a more elaborate description is needed for humans                 |
| forms       | mandatory | Not used for publishing TD with the Hub. (2)                              |  
| enum        | optional  | Restricted set of values                                                  |
| unit        | optional  | unit of the value                                                         |
| readOnly    | optional  | true for properties that are attributes, false for writable configuration |
| writeOnly   | optional  | not used. See above                                                       |
| default     | optional  | Default value to use if no value is provided                              |
| minimum     | optional  | type number/integer: Minimum range value for numbers                      |
| maximum     | optional  | type number/integer: Maximum range value for numbers                      |
| minLength   | optional  | type string: Minimum length of a string                                   |
| maxLength   | optional  | type string: Maximum length of a string                                   |                               

Notes:

1. Property names are standardized as part of the vocabulary so consumers can understand their
   purpose.
2. WoT specifies Forms to define the protocol for operations. In HiveOT all operations operate via a
   message bus with a simple address scheme. There is therefore no need for Forms. In addition,
   requiring a Forms section in every single property description causes unnecessary bloat that
   needs to be generated, parsed and stored by exposed and consumed things.
3. In HiveOT the namespace for properties, events and actions is shared to avoid ambiguity. A change
   in property value can lead to an event with the property name. Writing a property value is done
   with an action of the same name. (the WoT group position on this is unknown. Is this intended?)
4. The use of readOnly and writeOnly attributes is unfortunate as it is seems redundant but isn't.
   What does writeOnly true mean? HiveOT things only use 'readOnly'. When omitted it is
   assumed to be true. Since JSON doesn't support default values, it might cause parsing
   complications. HiveOT only uses readOnly and ignores writeOnly. readOnly false means writable.


## Events

Changes to the state of a Thing are published using Events. The TD describes the events that a Thing
publishes in its events affordance section and is serialized as JSON in the following format:

```
{
  "events": {
    "{name}": {
      ...InteractionAffordance,
      "data": {
        dataSchema
      },
      "dataResponse: {EventResponseData}"
    }
  }
}
```

Where:

* {name}: The name of the event. Event names share the same namespace as property names. The names
  are standardized in the HiveOT vocabulary.
* data: Defines the data schema of event messages. The content follows
  the [dataSchema](https://www.w3.org/TR/wot-thing-description11/#dataschema) format, similar to
  properties.
* dataResponse: Describes the data schema of a possible response to the event. EventResponses are
  currently not used in HiveOT.

The [TD EventAffordance](https://www.w3.org/TR/wot-thing-description11/#eventaffordance) also
describes optional subscription and cancellation attributes. These are not used in HiveOT as
subscription is not handled by a Thing but by the Hub API/MQTT message bus.

### The "properties" Event

In HiveOT, changes to property values are sent using events. Rather than sending a separate event for
each property, HiveOT defines a 'properties' event. This events contains a properties map with
property name-value pairs. The concern this tries to address is that this reduces the amount of
events that need to be sent by small devices, reducing battery power and bandwidth.

As the 'properties' event is part of the HiveOT standard does not have to be included in the '
events' section of the TDs (but is recommended).  (should this be part of a HiveOT @content? tbd)

```json
{
  "events": {
    "properties": {
      "data": {
        "title": "Map of property name and new value pairs",
        "type": "object"
      }
    }
  }
}
```

For example, when a name property has changed the event looks like:

```
{
  "name": "new name",
}
```

## Actions

Actions are used to control inputs and change the value of configuration properties.

The format of actions is defined in the Thing Description document
through [action affordances](https://www.w3.org/TR/wot-thing-description/#actionaffordance).

Note: The specification 'requires' a 'forms' element in each action affordance. HiveOT deviates from
the standard in that the 'forms' element is not used for individual actions, events and properties.
Instead, a single generic address format is used of "things/{id}/action/{name}". Ideally this
can be defined generically at the top level of the TD but no such specification exists at the time
of writing. This section might be revised in the future.

```
{
  "actions": {
    "{name}": ActionAffordance,
    ...
  }
}
```

Where {name} is the name of the action as defined in the vocabulary, and ActionAffordance describes
the action details. The action name shares the namespace with events and properties. The result of
an action can be notified using an event with the same name and shown with a property of the same
name:

```
{
  ...interactionAffordance,
  "input": {},
  "output": {},
  "safe": true
  |
  false,
  "idempotent": true
  |
  false
}
```

For example, the schema of an action to control an onoff switch might be defined as:

```
{
  "actions": {
    "onoff": {
      "title": "Control the on or off status of a switch",
      "input": {
        "type": "boolean",
        "title": "true/false for on/off"
      }
    }
  }
}
```

The action message to turn the switch on then looks like this:

```json
{
  "value": true
}
```

### The 'properties' Action

Similar to a properties Event, HiveOT standardizes a "properties" action. To change a property value, a properties action must be submitted containing a map of requested property name and value pairs. No additional action affordance is needed to write properties although this is recommended.

For example, when the Thing configuration property called "alarmThreshold" changes, the action looks
like this.

```json
{
  "alarmThreshold": 25
}
```

## Links

The spec describes a [link](https://www.w3.org/TR/wot-thing-description11/#link) as "A link can be
viewed as a statement of the form "link context has a relation type resource at link target", where
the optional target attributes may further describe the resource"

In HiveOT a link can be used as long as it is not served by the IoT device, as this would conflict
with the paradigm that "Things are not servers".

## Forms

The WoT specification for a [Form](https://www.w3.org/TR/wot-thing-description11/#form) says: "A form
can be viewed as a statement of "To perform an operation type operation on form context, make a
request method request to submission target" where the optional form fields may further describe the
required request."

As HiveOT does not allow direct protocol access to Things, Forms are ignored in published TDs. The Hub might instead replace the TD Forms section with a FOrm describing the Hub protocol to interact with the device via the Hub.

### SecuritySchema 'scheme' (1)

In HiveOT all authentication and authorization is handled by the Hub. Therefore, the security scheme
section only applies to Hub interaction and does not apply to HiveOT Things. The Hub service used to interact with Things will publish a TD that includes a SecuritySchema needed to interaction with the Hub.

# REST APIs

HiveOT compliant Things do not implement TCP/Web servers. All interaction takes place via HiveOT Hub services
and message bus. Therefore, this section only applies to Hub services that provide a web API. Instead the Hub Gateway Service provides web REST API's.

Hub services that implement a REST API follows the approach as described in Mozilla's Web Thing REST
API](https://iot.mozilla.org/wot/#web-thing-rest-api).

```http
GET https://address:port/things/{thingID}[/...]
```

The WoT examples often assume or suggest that Things are directly accessed, which is not allowed in HiveOT. Therefore, the implementation of this API in HiveOT MUST follow the following
rules:

1. The Thing address is that of the hub it is connected to.
2. The full thing ID must be included in the API. The examples says 'lamp' where a Thing ID is a
   URN: "urn:device1:lamp" for example.
