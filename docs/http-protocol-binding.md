# WoST Hub HTTP Protocol Binding - under construction

This document describes the HTTP API used to exchange messages between Exposed Things and Consumed Things.

It is based on the Dec 2020 draft specification of [WEB Thing API](https://webthings.io/api/) and the Form templates
proposed in [WoT Binding Templates](https://w3c.github.io/wot-binding-templates/#property-forms).

**The WoST-go library currently only supports the MQTT binding. This specification is for future reference.**

## Security Schemes

Where applicable, the Hub supports the API's with [security schemes](https://www.w3.org/TR/wot-thing-description/):

This service supports :

* BearerSecurityScheme with the access token obtained

A valid client certificate signed by the Hub CA can also be used.

### Provisioning

Before a device can exchange Thing information with the Hub and its consumers it must be provisioned. Provisioning means
exchanging credentials for secure connectivity and identification. The credentials are defined by the security schemes.

The provisioning API is defined in the 'idprov' specification.

### Publish Thing Description Document

Exposed Things publish their Thing Description Document to subscribers.

> ### Request
> ```https
> HTTP PUT https://{hub}/things/{thingID}/td
> Accept: application/json
> {
>    ...Full TD...
> }
> ```
> ### Response
> ```https
> 200 OK
> ```

Where

* {hub} is the DNS name or IP address of the Hub
* {thingID} is the ID of the thing

### Get Thing Property Values

This returns the property values of a Thing from the Hub shadow registry.
TD's published via this API will have forms added that describe how to access Thing properties via this API.
See [Property Forms](https://w3c.github.io/wot-binding-templates/#property-forms) for details.


> #### request
> ```http
> HTTP GET https://{hub}/things/{thingID}/action/readproperty|readallproperties|readmultipleproperties
> Accept: application/json
> ```
> ### Response
> ```json
> {
>   "property1": "value 1",
>   ...
> }
> ```

### Write Property Values

Consumed Things can request an update to Thing properties by sending a write action request to an Exposed Thing.

The property {value} schema is described by the corresponding property affordance in the Thing Description Document.

> ### Request
> ```http
> HTTP PUT https://{hub}/things/{thingID}/action/{propertyName}
> Accept: application/json
> {value}
> ```
> ### Response
> ```http
> 200 OK

### Publish Thing Events

Exposed Things notify Consumed Things of an event that has happened.

The event {value} schema is described by the corresponding event affordance in the Thing Description Document.

Note the shared namespace with property names. If the name of a property is used then the property affordance is used
instead of the event affordance.


> ### Request
> ```http
> PUT /things/{thingID}/events/{eventName}
> Accept: application/json
> Content:
> {data}
> ### Response
> ```http
> 200 OK

### Request Action

Consumed Things can request an action on a device from an Exposed Thing.

The action {data} schema is described in the corresponding action affordance in the Thing Description Document.

> ### Request
> ```http
> PUSH /things/{thingID}/action/{actionName}
> Accept: application/json
> Content:
> {data}
> ### Response
> ```http
> 200 OK

## API comparison

The HTTP API is intended to conform to the WoT API standard. As this is not yet defined this is a best guess.
