
# WoST Hub API

This document describes the API's provided by the WoST Hub and its core plugins.

It is based on the Dec 2020 draft specification of [WEB Thing API](https://webthings.io/api/). 

Note that this is an abbreviated description of the full API. The full API will be defined in Swagger in the future.

## Security Schemes

Where applicable, the Hub supports the API's with [security schemes](https://www.w3.org/TR/wot-thing-description/): 
* BasicSecurityScheme
* DigestSecurityScheme
* APIKeySecurityScheme
* BearerSecurityScheme
* PSKSecurityScheme 
* OAuth2SecurityScheme

Clients can connect to the API using one of these schemes.


## HUB HTTP API 

This HUB HTTP API is an interface that lets Things send updates and events and answers to request for Thing information.

* Things can publish updates to their TD in the Hub shadow registry
* Things can publish updates to their property values 
* Things can publish events
* Things can retrieve property value update requests made in the last 24 hours
* Things can retrieve action requests made in the last 24 hours
* Consumers can publish action requests
* Consumers can publish thing property update requests
* Consumers can request a thing's TD
* Consumers can request a thing's last reported property values

This interface has the following limitations:
* This API is not a Directory Service. Consumers cannot query TDs. 
* Thing property values cannot be requested.
* Thing events cannot be requested.

See also the MQTT and WebSocket API's for a live stream of information between Things and Consumers.


### Provisioning

When a Thing is provisioned by the Hub, they exchange credentials for secured connectivity and message exchange. The credentials are defined by the security schemes.

> ### This section is still to be defined

### Get a Thing Description
=== FUTURE ===
This returns the most receive Thing Description from the shadow registry

> ### Request 
> ```http
> HTTP Get https://{hub}/things/{id}
> Accept: application/json
> ```

> ### Response
> 200 OK
> ```json
> {
>    Full TD
> }
> ```

> ### Response
> 404 NOT FOUND
> ```

### Update a Thing Description Document

This updates a Thing Description Document in the shadow registry.

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


### Get All Thing Property Values

==Future==

This returns the property values of a Thing from the shadow registry.

> #### request
> ```http
> HTTP GET https://{hub}/things/{id}/properties
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


## MQTT API

The MQTT API is intended for live communication between Things and Consumers. This API does not support request or queries for Thing Descriptions.

### Provisioning

> ### TBD


### Get a Thing Description Document

Not supported. Subscribe to updates from a Thing or use the Directory service. 

N/A

### Update a Thing Description Document

This updates a Thing Description Document in the shadow registry.

Thing publishes an updated TD
> Topic: things/{id}
> Content: 
> ```json
> {
>   Full TD
> }
>```

### Get A Single Thing Property Value 

Not supported. Subscribe to updates from a Thing or use the Directory service. 

> N/A

### Get All Thing Property Values

Not supported. Subscribe to updates from a Thing or use the Directory service. 

> N/A

### Update Of A Thing Property Values 

This updates Thing property values in the shadow registry. 

> Topic: things/{id}/properties
> Content: 
> ```json
> {
>    "{property1}": {value1}
>    ...
> }
>```

### Set Thing Property Values

This requests that a Thing updates its property value. This does not update the shadow registry.

> Topic: things/{id}/set
> Content: 
> ```json
> {
>    "{property1}": {value1}
>    ...
> }
>```


### Publish Thing Events

This notifies subscribers of an event that happened on a Thing.

> Topic: things/{id}/events
> ```json
> Content:
> {
>   "event1": {
>     "data": {value},
>     "timestamp": {iso8601 timestamp},
>   },
>    ...
> }
> ```


### Request A Thing Performs An Action

> Topic: things/{id}/action
> ```json
> Content:
> {
>   "event1": {
>     "data": {value},
>     "timestamp": {iso8601 timestamp},
>   },
>    ...
> }
> ```



## WebSocket API

The WebSocket API is an alternative pub/sub message bus interface that is built-in into the Hub. It is automatically started when the Hub is started. The Hub configuration must select whether the MQTT or the WebSocket API is used. It cannot use both.

The addressing and message content are the same as the MQTT API.





## API comparison

The Thing API conforms to the WoT API standard. As this is not yet defined this is a best guess.

Three similar protocols can be used by Things to publish and update their TD, publish events and receive actions: MQTT, HTTP and WebSocket, all over TLS.

The message content is JSON encoded and contains a Thing Description Document (TD), a property fragment, an event or an action.


| Message                   | HTTP Address                       | MQTT Address           | WebSocket Address      |
| ------------------------- | ---------------------------------- | ---------------------- | ---------------------- |
| Get TD                    | GET /things/{id}                   | N/A                    | N/A                    |
| Update of TD              | PUT /things/{id}                   | things/{id}            | things/{id}            |
| Get property value        | GET /things/{id}/properties/{name} | N/A                    | N/A                    |
| Get all property values   | GET /things/{id}/properties        | N/A                    | N/A                    |
| Update of property values | PUT /things/{id}/properties        | things/{id}/properties | things/{id}/properties |
| Set property value        | PUT /things/{id}/set               | things/{id}/set        | things/{id}/set        |
| Publish events            | PUT /things/{id}/events            | things/{id}/events     | things/{id}/events     |
| Receive actions           | GET /things/{id}/actions           | N/A                    | N/A                    |
| Request a Thing action    | PUT /things/{id}/action            | things/{id}/action     | things/{id}/action     |
|                           |
