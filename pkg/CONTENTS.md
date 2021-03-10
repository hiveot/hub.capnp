The pkg folder contains the client side libraries for building hub plugins.

 ./certs      Certificate generation functions. Used by the server
 ./config     Configuration file and commandline parser. Convenience for writing plugins
 ./logging    Setup of plugin logging. Convenience for writing plugins 
 ./messaging  Client library for connecting to the message bus using the message bus, publish messages and subscribe to channels

The built-in smbus message bus is a simple websocket based message bus for plugins to publish and subscribe to hub channels. It is protected with TLS client and server certificates and can be used to connect plugins on the local area network to the hub. Most commonly the plugins will be installed on the same host as the message bus server itself. See the smbserver plugin for the server code.

