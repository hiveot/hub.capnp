# WoST Hub MQTT Protocol Binding

This document describes the MQTT protocol used to exchange messages between Exposed Things and
Consumed Things.

## Security Schemes

WoST uses the NoSecurityScheme for Things.

Clients authenticate with the MQTT message bus with one of the following methods:

* client certificate: A valid client certificate, signed by the Hub CA, provide authentication and
  authorization based
  on the certificate OU. This is intended for Thing devices and administrators.
* login name and access token: The access token is obtained from the authentication service by
  providing the user
  password.

Once authenticated to the message bus, the message bus handles authorization based on the client's
role and group.

## MQTT Messages

Interaction between exposed things and consumed things takes place via MQTT messages as described in
the following
sections.

### Publish Thing Description Document

Exposed Things publish their Thing Description Document to subscribers on startup and on changes.
This notification can
be sent at any time.

> Topic: things/{thingID}/td
> Content:
> ```json
> {
>   ...Full TD...
> }
>```

Consumed Things subscribe to this topic to receive update to the TD.

### Notify Of Change To Property Values

When Exposed Things update their property values this is sent as an event. The {value} payload
schema is described in the property affordance of the Thing Description document. No event
affordance is needed for properties.

> ### Update of a single property
> Topic: things/{thingID}/event/{propertyName}
>
> Contains the value of the changed TD property:
> ```json
>    {value}
>```

For efficiency reasons it is possible to include multiple properties in one event. The payload is a
property map containing the property values. The topic uses the prescribed event name 'properties'.


> ### Update of multiple properties
> Topic: things/{thingID}/event/properties
>
> Contains a map of property name-value pairs of changed TD properties:
> ```json
> {
>    "{propertyName}": {value},
>    ...
> }
>```

Consumed Things subscribe to these topics to receive updates to property values as
they are published.

### Write Property Values

Consumed Things can request an update to Thing properties by sending a write action request to an
Exposed Thing.

The property {value} schema is described by the corresponding property affordance in the Thing
Description Document.

> Topic: things/{thingID}/action/{propertyName}
> Content:
> ```json
> {value}
>```

### Publish Thing Events

Exposed Things notify Consumed Things of an event that has happened.

The event {value} schema is described by the corresponding event affordance in the Thing Description
Document.

Note the shared namespace with property names. If the name of a property is used then no event
affordance is needed.

> ### Events Message
> Topic: things/{thingID}/event/{eventName}
>
> Contains the value of the TD 'event' object as defined in the TD:
>```json
>   {value}
>```

### Request Action

Consumed Things can request an action on a device from an Exposed Thing.

The action {data} schema is described in the corresponding action affordance in the Thing
Description Document.

> Topic: things/{thingID}/action/{actionName}
> ```json
> Content:
> {data}
> ```
