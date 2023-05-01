# Hub MQTT Gateway

This is a workaround for the lack of a mature javascript capnproto client. In addition it can serve as an integration tools for clients that are mqtt-empowered.

## Status

This service is in development


## Summary

Javascript clients do not have an easy way to establish a capnproto connection. This service provides a MQTT gateway to the hub, offering limited capabilities for use by javascript clients running in a web browser, or any other MQTT clients.

Features:
1. User based authentication
2. Subscribe to TD, events and actions
3. Publish TD, actions and events
4. Read directory [1] 
5. Read history [1]

As mqtt is a pub/sub protocol, not a request/response protocol. In order to handle request/response communication, the service supports request and response topics.
When a valid request is received on the 'hiveot/request' topic, the response is published on the 'hiveot/response/clientID' topic.


## Mochi

This service is built around the [mochi embedded MQTT broker](https://github.com/mochi-co/mqtt)

