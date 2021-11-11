
# WoST Hub HTTP API

This document describes the HTTP API used to connect to the WoST Hub. The server side is implemented by the Hub HTTP protocol binding.

It is based on the Dec 2020 draft specification of [WEB Thing API](https://webthings.io/api/) and the Form templates proposed in [WoT Binding Templates](https://w3c.github.io/wot-binding-templates/#property-forms).

Note that the WoT specification is still in draft. Also note that the specification has some complicated ways of doing simple things. This document includes some proposals to keep things simple. The WoST protocol binding might end up having to implement the simple form and the complicated form once the standard is finalized.


## Security Schemes

Where applicable, the Hub supports the API's with [security schemes](https://www.w3.org/TR/wot-thing-description/): 
* BasicSecurityScheme
* DigestSecurityScheme
* APIKeySecurityScheme
* BearerSecurityScheme
* PSKSecurityScheme 
* OAuth2SecurityScheme

The intent is that clients can connect to the API using one of these schemes.


### Signing and Encryption

The WoT HTTP protocol binding does not specify how to sign and encrypt messages between Consumer and Thing (via the Hub).

The Hub intents to support identity verification through JWS. A Hub that has enabled identity verification requires that all messages are published with a JWS signature using the public key that the Hub provided during provisioning. Consumers can verify the authenticity of the sender.

Last, for messages with sensitive information, a publisher can choose to encrypt the message payload using JWE and the public key of the intended receiver. 

The protocol binding for signing and encryption of messages over MQTT is not specified in the WoT standard. WoST applies this extension as optional to allow for backwards compatibility. See the Message format section for details.



## HUB HTTP API Methods

This HUB HTTP API is an interface that lets Things send updates and events and answers to request for Thing information.

* Things can publish updates to their TD to the Hub
* Things can publish updates to their property values 
* Things can publish events
* Things can retrieve configuration update requests made in the last 24 hours (future)
* Things can retrieve action requests made in the last 24 hours (future)
* Consumers can request property values via the Hub
* Consumers can publish action requests via the Hub
* Consumers can publish thing property update (configuration update) requests via the Hub


### Provisioning

Before a device can exchange Thing information with the Hub and its consumers it must be provisioned. Provisioning means exchanging credentials for secure connectivity and identification. The credentials are defined by the security schemes. 

> ### This section is still to be defined

### Update a Thing Description Document

Thing sends an update of the a Thing Description Document.

> ### Request 
> ```http
> HTTP PUT https://{hub}/things/{id}
> Accept: application/json
> {
>    Full TD
> }
> ```
> ### Response
> ```http
> 200 OK
> ```

Where
* {hub} is the DNS name or IP address of the Hub
* {id} is the ID of the thing


### Get Thing Property Values

==Future==

This returns the property values of a Thing from the Hub shadow registry.
TODO: Use the [Property Forms](https://w3c.github.io/wot-binding-templates/#property-forms) to advertise how to get the property values.

TBD: Note: This seems unnecesary complex. The readOnly/writeOnly requirements mentioned is just asking for confusion. Also having multiple methods for the same thing like readproperty, readallproperties, readmultipleproperties is unnecesary.

Proposal 1: Change the readOnly/writeOnly attributes to 'writable'. readonly is implied.
Proposal 2: Have only a single opvalue for readproperty. Return the properties (one or more) specified. If no properties are specified then return all.
Proposal 3: Have only a single opvalue for writeproperty. Simply write the properties specified.

> #### request
> ```http
> HTTP GET https://{hub}/things/{id}/values
> Accept: application/json
> ```
> ### Response
> ```json
> {
>   "property1": "value 1",
>   ...
> }
> ```

### Get A Single Thing Property Value

==Future==

This returns the property value of a Thing from the shadow registry.

> #### request
> ```http
> HTTP GET https://{hub}/things/{id}/properties/{name}
> Accept: application/json
> ```
> ### Response
> ```json
> {
>   "{name}": "value",
> }
> ```

### Update Of A Thing's Property Values

This updates Thing property values in the shadow registry. Only the properties that are updated need to be included. 

> ### Request
> ```http
> HTTP PUT https://{hub}/things/{id}/properties
> Accept: application/json
> {
>    "{property1}": {value1},
>    ...
> }
> ```
> ### Response
> ```http
> 200 OK
> {
>    "{name}": {value},
>    ...
> }
### Set Thing Property Values

==Future==

This requests that a Thing updates its property value

> ### Request
> ```http
> HTTP PUT https://{hub}/things/{id}/set
> Accept: application/json
> {
>    "{property1}": {value1},
>    ...
> }
> ```
> ### Response
> ```http
> 200 OK
> {
>    "{name}": {value},
>    ...
> }

### Publish Thing Events

This notifies subscribers of an event that happened on a Thing.

> ### Request
> ```http
> PUT /things/{id}/events
> Accept: application/json
> Content:
> {
>   "event1": {
>     "data": {value},
>     "timestamp": {iso8601 timestamp},
>   },
>    ...
> }
> ### Response
> ```http
> 200 OK
> {
>    {event}
> }


### Request Queued Actions 

==Future==

This returns the queued actions of a Thing from the shadow registry.

This is intended for Things to request the actions that have been queued since their last request. After the result is returned the queue is emptied. It is intended for Things that cannot use the WebSocket or MQTT API.
This call is not idempotent.

> ### Request
> ```http
> Get /things/{id}/actions
> Accept: application/json
> 
> ### Response
> ```http
> 200 OK
> Content:
> {
>   "{action1}": {
>     "input": {
>       "{param1}": {value},
>       "{param2}": {value},
>     },
>   },
>    ...
> }



## API comparison

The HTTP API is intended to conform to the WoT API standard. As this is not yet defined this is a best guess.

