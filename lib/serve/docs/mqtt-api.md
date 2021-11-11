
# WoST Hub MQTT API

This document describes the MQTT API used to connect to the WoST Hub. The server side is provided by the Hub MQTT protocol binding.

It is based on the Dec 2020 draft specification of [WEB Thing API](https://webthings.io/api/) with some changes to make it WoST compatible.

Note that this is an abbreviated description of the full API. The full API will be defined in Swagger in the future.

## Security Schemes

Where applicable, the Hub supports the API's with [security schemes](https://www.w3.org/TR/wot-thing-description/): 
* BasicSecurityScheme   (todo)
* DigestSecurityScheme  (todo)
* APIKeySecurityScheme  (todo)
* BearerSecurityScheme  (todo)
* PSKSecurityScheme     (todo)
* OAuth2SecurityScheme  (todo)

Clients can connect to the API using one of these schemes. 

Support for identity verification is provided through JWS. A Hub that has enabled identity verification requires that all messages are published with a JWS signature using the public key that the Hub provided during provisioning. Consumers can verify the authenticity of the sender.

Last, for messages with sensitive information, a publisher can choose to encrypt the message payload using JWE and the public key of the intended receiver. 

The protocol binding for signing and encryption of messages over MQTT is not specified in the WoT standard. WoST applies this extension as optional to allow for backwards compatibility. See the Message format section for details.

## HUB MQTT API Messages 

The MQTT API is intended for live communication between Things, Services and Consumers. 

* Things publishes updates to their TD 
* Things publishes updates to their property values 
* Things publishes events
* Consumer publishes a configuration update requests
* Consumer publishes an action request

### Provisioning

A WoST compatible device must be provisioned using one of the client API's. When a device is provisioned by the Hub, they exchange credentials for secured connectivity and message exchange. The credentials are defined by the security schemes. Note that a device can manage multiple Things.


> #### TBD


### Thing Publishes a Thing Description Document

Thing publish a Thing Description Document to subscribers

> Topic: things/{id}/td
> Content: 
> ```json
> {
>   Full TD
> }
>```

Consumers can subscribe to this topic or the 'things/+/td' topic to receive TDs as they are published.


### Thing Publishes Update To Property Values

**Note1: The WoT TD specification does not differentiate between sensor/actuator, configuration and status properties. There are several options to address this:**
* Option 1. Each property has an attribute describing the property category: sensor, actuator, status, configuration
* Option 2. Use only 4 properties in the WoST TD: sensors, actuators, status, configuration. Each of these properties is an object that contains properties for their respective attributes.

Since option 1 keeps the API the smallest. The solution chosen is to use the 'category' attribute in a property definition and not to add separate API's for sensors, actuators, configuration, and status.

Notify consumers that one or more Thing property values are updated. 

> ### Values Message
> Topic: things/{id}/values
> 
> Contains the values of changed TD properties:
> ```json
> {
>    "{property1}": {value1}
>    ...
> }
>```

Consumers can subscribe to a this topic or the 'things/+/values' wildcard topic to receive updates to property values as they are published.

### Thing Publishes Events

Notify consumers of one or more events that have happened on a Thing.

> ### Events Message
> Topic: things/{id}/events
> 
> Contains the values of the TD 'events' object properties:
>```json
>{
>  "event1": {
>    "value": {value},
>    "timestamp": {iso8601 timestamp},
>  },
>   ...
>}
>```

### Consumer Request Updates To Thing Configuration Values

Consumer request that a Thing updates its configuration property value(s). Note that actuator values are updated through actions.

Things subscribe to this address to receive the update request. If successful this results in a publication of a configuration property values update message by the Thing.

> Topic: things/{id}/config
> Content: 
> ```json
> {
>    "{property1}": {value1}
>    ...
> }
>```

### Consumer Publishes Request That A Thing Performs An Action

Consumer requests an action on a Thing. 

Things subscribe to this address to receive the action requests. If successful this results in a publication of an actuator property value update message by the Thing.

> Topic: things/{id}/action
> ```json
> Content:
> {
>   "action1": {
>     "data": {value},
>     "timestamp": {iso8601 timestamp},
>   },
>    ...
> }
> ```


