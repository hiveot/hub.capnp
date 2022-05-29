# WoST use of the TD - under development

WoST intends to be compliant with the WoT Thing Description specification.
The latest known draft at the time of writing
is [Mar 2022](https://www.w3.org/TR/wot-thing-description11/#thing)

Interpreting this specification is done as a best effort. In case where discrepancies are reported
they will be corrected when possible as long as they do not conflict with the WoST core paradigm
of 'Things do not run servers'.

The WoT specification is closer to a framework than an application. As such it doesn't dictate how
an application should use it. This document describes how the WoST information model and behavior
maps to the WoT TD.

# WoST IoT Device Model

## I/O Devices, Gateways and Publishers are IoT 'Things'

Most IoT devices are pieces of hardware that have embedded software that manages its behavior.
Virtual IoT devices are build with software only but are otherwise considered identical to hardware
devices.

IoT devices often fulfill multiple roles: a part provides network access, a part provides access to
inputs and outputs, a part reports its state, and a part that manages its configuration.

WoST makes the following distinction based on the primary role of the device. These are identified
by their device type:

* A gateway is a Thing that provides access to other independent Things. A Z-Wave controller
  USB-stick is a gateway that uses the Z-Wave protocol to connect to I/O devices. A gateway is
  independent of the Things it provides access to and can have its own inputs or outputs. Gateways
  are often used when integrating with non-WoST Things.
* A publisher is a Thing that publishes other Thing information to the WoST Hub. Publishers have
  their own ID that is included as part of the Thing ID of all Things that it publishes. A publisher
  has by default authorization to publish and subscribe to the things it is the publisher of.
  Publishers often are services that convert between the WoT/WoST standards and a native protocol.
* An I/O device is Thing whose primary role is to provide access to inputs and outputs and has its
  own attributes and configuration. In case of hybrid hardware where attributes and configuration
  are managed by a parent device then the inputs and outputs are also considered to be part of the
  parent device.
* A Hub bridge is a device that connects two Hubs and shares Thing information between them.

## Thing Description Document (TD)

The Thing Description document is
a [W3C WoT standard](https://www.w3.org/TR/wot-thing-description11/#thing) to describe Things. TDs
that are published on the Hub MUST adhere to this standard and use the JSON representation format.

The TD consists of a set of attributes and properties that WoST uses to describe WoST things and
their capabilities.

## TD Attributes

TD Attributes that WoST uses are as follows. Attributes marked here as optional in WoT are
recommended in WoST:

| name               | mandatory | description                                    |
|--------------------|-----------|------------------------------------------------|
| @context           | mandatory | "http://www.w3.org/ns/td"                      |
| id                 | mandatory | Unique Thing ID following the WoST ID format   |
| title              | mandatory | Human readable description of the Thing        |
| modified           | optional  | ISO8601 date this document was last updated    |
| properties         | optional  | Attributes and Configuration objects           |
| version            | optional  | Thing version as a map of {'instance':version} |
| actions            | optional  | Actions objects supported by the Thing         |
| events             | optional  | Event objects as submitted by the Thing        |
| security           | mandatory | Protocol dependent security (see note1)       |
| securityDefinition | mandatory | Protocol dependent security (see note1)       | 

note1: Consumers do not connect directly to the IoT device and authentication & authorization is
handled by the Hub services. As a result the security definitions in the TD depend on the method
used to access the WoST service, not the IoT device.

* WoST compatible IoT devices can simply use the 'nosec' security type when creating their TD and
  use a NoSecurityScheme as securityDefinition.
* Consumers, which access devices via 'Consumed Things', only need to know how connect to the Hub
  services and MQTT message bus. No knowledge of the IoT device protocol is needed.

## WOST Thing ID Format

All Things have an ID that is unique to the Hub. To support uniqueness, authorization, and sharing
with other Hubs, a WoST ID is constructed as follows:

> WoST compatible IoT devices publish their 'Thing' with the following ID:
> urn:{zone}/{deviceID}/{deviceType}

> Protocol binding or services that publish multiple 'Things', for example a ZWave protocol binding,
> include the service as the publisherID using the following format:
> urn:{zone}/{publisherID}/{deviceID}/{deviceType}

* urn: prefix defined by the WoT standard
* {zone} is the zone where the thing originates. All local things are in the 'local' zone. A bridge
  that shares Things from another Hub will change the zone of of the shared Things to that of the
  bridge domain.
* {publisherID} is the deviceID of the IoT device that provides access to a Thing. It must be unique
  on the Hub it is publishing to. Publishers MUST be WoST compliant and implement the WoT/WoST
  standard.
  IoT devices that publish their own Thing can omit the publisher. In that case the publisherID and
  deviceID are the same and must be unique on the Hub.
* {deviceID} is the ID of the hardware or software Thing being accessed. It must be unique on the
  publisher that is publishing it. In case of protocol bindings this is the ID of the original
  protocol.
  In case of a WoST compatible IoT Device this can be a mac address or other locally unique feature
  of the device.
* {deviceType} is the type of Thing as defined in
  the [WoST vocabulary](https://github.com/wostzone/wost-go/blob/pkg/vocab/IoTVocabulary.go). See
  the constants with prefix DeviceType. It describes the primary role of the device.

When integrating with 3rd party systems that use a URI as the ID, the ID can be used as-is. If
the ID is not a URI then it must be used as the deviceID, while the publisherID is that of the
service that provides the protocol binding. Using a 3rd party ID as-is can lead to reduced
capabilities for bridging, queries in the directory and other services.

The Thing ID is created by the publisher of a Thing.

## Thing Properties

Thing Properties describe the Thing and the state it is in. For example, device type and version are
properties. Read-only properties are considered attributes while writable properties are
configuration.

The WoT TD describes properties with
the [PropertyAffordance](https://www.w3.org/TR/wot-thing-description11/#propertyaffordance). This is
a sub-class of
an [interaction affordance](https://www.w3.org/TR/wot-thing-description11/#interactionaffordance)
and [dataschema](https://www.w3.org/TR/wot-thing-description11/#dataschema).

WoST uses the following attributes to describe properties.

| Attribute   | WoT       | description                                                               |
|-------------|-----------|---------------------------------------------------------------------------|
| name        | optional  | Name used to identify the attribute in the TD Properties object. (1)      |
| type        | optional  | data type: string, number, integer, boolean, object, array, or null       |
| title       | optional  | Human description of the attribute.                                       |
| description | optional  | In case a more elaborate description is needed for humans                 |
| forms       | mandatory | Tbd. WoST uses a standard MQTT address format for all operations          | 
| value       | optional  | Value of the attribute.                                                   |
| minimum     | optional  | Minimum range value for numbers                                           |
| maximum     | optional  | Maximum range value for numbers                                           |
| enum        | optional  | Restricted set of values                                                  |
| unit        | optional  | unit of the value                                                         |
| readOnly    | optional  | true for properties that are attributes, false for writable configuration |
| writeOnly   | optional  | not used. See above                                                       |
| default     | optional  | Default value to use if no value is provided                              |

Notes:

1. Property names are standardized as part of the vocabulary so consumers can understand their
   purpose.
2. WoT specifies Forms to define the protocol for operations. In WoST all operations operate via a
   message bus with a simple address scheme. There is therefore no need for Forms. In addition,
   requiring a Forms section in every single property description causes unnecessary bloat that
   needs to be generated, parsed and stored by exposed and consumed things.
3. In WoST the namespace for properties, events and actions is shared to avoid ambiguity. A change
   in property value can lead to an event with the property name. Writing a property value is done
   with an action of the same name. (the WoT group position on this is unknown. Is this intended?)
4. The use of readOnly and writeOnly attributes is unfortunate as it is seems redundant but isn't.
   What does writeOnly true mean? WoST things only use 'readOnly'. When omitted it is
   assumed to be true. Since JSON doesn't support default values, it might cause parsing
   complications. WoST only uses readOnly and ignores writeOnly. readOnly false means writable.

## Events

Changes to the state of a Thing are published using Events. The TD describes the events that a Thing
publishes in its events affordance section and is serialized as JSON in the following format:

```json
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
  are standardized in the WoST vocabulary.
* data: Defines the data schema of event messages. The content follows
  the [dataSchema](https://www.w3.org/TR/wot-thing-description11/#dataschema) format, similar to
  properties.
* dataResponse: Describes the data schema of a possible response to the event. EventResponses are
  currently not used in WoST.

The [TD EventAffordance](https://www.w3.org/TR/wot-thing-description11/#eventaffordance) also
describes optional subscription and cancellation attributes. These are not used in WoST as
subscription is not handled by a Thing but by the MQTT message bus.

### The "properties" Event

In WoST, changes to property values are sent using events. Rather than sending a separate event for
each property, WoST defines a 'properties' event. This events contains a properties map with
property name-value pairs. The concern this tries to address is that this reduces the amount of
events that need to be sent by small devices, reducing battery power and bandwidth.

As the 'properties' event is part of the WoST standard does not have to be included in the '
events' section of the TDs (but is recommended).  (should this be part of a WoST @content? tbd)

Alternative (tbd) each property change is a separate event. The concern is that this can lead to a
lot of events which can consume significant resources on small devices.

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

For example, when a temperature has changed to 21 degrees and humidity to 55%, the event payload
looks like this.

```json
{
  "temperature": "21",
  "humidity": "55"
}
```

## Actions

Actions are used to control inputs and change the value of configuration properties.

The format of actions is defined in the Thing Description document
through [action affordances](https://www.w3.org/TR/wot-thing-description/#actionaffordance).

Note: The specification 'requires' a 'forms' element in each action affordance. WoST deviates from
the standard in that the 'forms' element is not used for individual actions, events and properties.
Instead, a single generic mqtt address format is used of "things/{id}/action/{name}". Ideally this
can be defined generically at the top level of the TD but no such specification exists at the time
of writing. This section might be revised in the future.

```json
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

```json
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

```json
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
  true
}
```

### Writing Properties Using Actions

As properties, actions and events share the same namespace. To write properties an action can be
used. As properties are already defined in the TD, no additional action affordance is needed to
write properties.

For example, when the Thing configuration property called "alarmThreshold" changes, the action looks
like this.

```json
{
  25
}
```

## Links

The spec describes a [link](https://www.w3.org/TR/wot-thing-description11/#link) as "A link can be
viewed as a statement of the form "link context has a relation type resource at link target", where
the optional target attributes may further describe the resource"

In WoST a link can be used as long as it is not served by the IoT device, as this would conflict
with the paradigm that "Things are not servers".

## Forms

The WoT specification for a [Form](https://www.w3.org/TR/wot-thing-description11/#form) says: A form
can be viewed as a statement of "To perform an operation type operation on form context, make a
request method request to submission target" where the optional form fields may further describe the
required request.

The provided example shows an HTTP POST to write a property.

In WoST an important constraint is that operations that interact with the Thing use the MQTT
protocol and topic of the Hub message bus as Things can only interact via the message bus.

Forms can use other protocols to describe interaction with external services. For example, to read
Thing property values, a form can define the https protocol to access a directory service that
collects the latest value by listening on the message bus. This is the approach taken by the WoST
directory service. In this example, the directory service can augment the TD of the thing to include
a form for the readproperties operation including its own endpoint.

This use of forms is still subject to changes in the future. Specifically, the use of a generic top
level form that can be applied to properties, events and actions is needed but not defined at the
time of writing.

### SecuritySchema 'scheme' (1)

In WoST all authentication and authorization is handled by the Hub. Therefore, the security scheme
section only applies to Hub services and does not apply to WoST Things. Things have a '
NoSecurityScheme' as they cannot be directly interacted with.

# REST APIs

WoST compliant Things do not implement servers. All interaction takes place via WoST Hub services
and message bus. Therefore, this section only applies to Hub services that provide a web API. For
example, the Directory Service and Provisioning Service provide web REST API's.

Hub services that implement a REST API follows the approach as described in Mozilla's Web Thing REST
API](https://iot.mozilla.org/wot/#web-thing-rest-api).

```http
GET https://address:port/things/{thingID}[/...]
```

Note 1: the WoT examples often assume or suggest that Things are directly accessed, which
is not allowed in WoST. Therefore, the implementation of this API in WoST MUST follow the following
rules:

1. The Thing address is that of the hub it is connected to.
2. The full thing ID must be included in the API. The examples says 'lamp' where a Thing ID is a
   URN: "urn:local:device1:lamp" for example.
