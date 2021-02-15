This folder contains the gateway server code and the built-in message bus server and client (bimb)

The built-in message bus is a simple websocket based message bus for plugins to publish and subscribe to gateway channels. It is protected with TLS client and server certificates and can be used to connect plugins on the local area network to the gateway. Most commonly the plugins will be installed on the same host as the message bus server itself.

