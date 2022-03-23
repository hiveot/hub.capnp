# MQTT protocol binding

This document specifies the MQTT protocol binding for the WoST Hub.

Bindings are based on the [W3C WoT bindings specification](https://w3c.github.io/wot-binding-templates/#creating-a-new-protocol-binding).

Documentation of a binding should contain:
* URI schema
* Mapping to WoT operations, eg readproperty, writeproperty, invokeaction, ...
* Document that specifies the protocol.



## URI Schema

The MQTT protocol is identified with the 'mqtt://' URI schema.


## WoT Operations Mapping

Copied verbatim from https://w3c.github.io/wot-binding-templates/bindings/protocols/mqtt/index.html. It is not clean how or where this is helpful. The MQTT protocol binding implements the MQTT message for each operation as described below. 


| operation               | binding                                  |
|-------------------------|------------------------------------------|
| readproperty            | "mqv:controlPacketValue": "SUBSCRIBE"    |
| writeproperty           | 	"mqv:controlPacketValue": "PUBLISH"     |
| observeproperty         | 	"mqv:controlPacketValue": "SUBSCRIBE"   |
| unobserveproperty       | 	"mqv:controlPacketValue": "UNSUBSCRIBE" |
| invokeaction            | 	"mqv:controlPacketValue": "PUBLISH"     |
| subscribeevent          | 	"mqv:controlPacketValue": "SUBSCRIBE"   |
| unsubscribeevent        | 	"mqv:controlPacketValue": "UNSUBSCRIBE" |
| readallproperties       | 	"mqv:controlPacketValue": "SUBSCRIBE"   |
| writeallproperties      | 	"mqv:controlPacketValue": "PUBLISH"     |
| readmultipleproperties  | 	"mqv:controlPacketValue": "SUBSCRIBE"   |
| writemultipleproperties | "mqv:controlPacketValue": "PUBLISH"      |

## Messages

### Event Message

topic: things/{thingID}/event/{eventName}

data: DataSchema as per EventAffordance


## Options

MQTT supports the following options. Note that these might not be supported in the current implementation. 

### To Retain or Not To Retain

Most MQTT implementations support 'retained' messages. The last received message on a topic is stored and after subscribing to a topic, this last message for this topic is immediately received. It acts as a cache. 

The plus of enabling retain is that the most recent TD's and events will be received immediately on connecting to the message bus.

The downside is that this can lead to a lot of messages when using wildcard subscription. Not all clients handle the avalanche of messages gracefully. It can also cause significant and costly bandwidth consumption.

In WoST the recommendation for consumers is NOT to use retainment unless there is a specific use-case to do so.  

### qos 

MQTT supports QOS of 0, 1 or 2. In WoST a default QOS of 1 is assumed (guaranteed delivery at least once).

A Qos of 0 can be used in case of high frequency updates of the same event, where intermittently dropped messages have little impact on the application.

For actions that are not idempotent a QOS of 2 (exactly once) can be used.

For example (not really much of an idea how this should look like):
```json
  "form": {
    "options": [
      "qos": "1"       
    ]
  }
```

### dup

This flag is set by the protocol binding if a received message is marked as a duplicate by the MQTT broker.

It is currently ignore. 

