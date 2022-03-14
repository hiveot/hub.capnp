# Mosquitto Manager 

This plugin manages the Mosquitto configuration, a lightweight and powerful MQTT message bus broker. In order to use this plugin the mosquitto broker must be installed.

MQTT is the primary communication method for the Hub and its plugins. Things and consumers can connect via MQTT to publish and subscribe to TD's, Events and Actions as described in the protocol binding. 

## Status

Status: early alpha. 

What it does:
- configure and launch a Mosquitto instance using the hub configuration
- configure mosquitto logging
- provide authentication using client certificates and username/password via the authentication plugin
- provide group/role authorization using the included ACL plugin

### Known Issues

Firefox connections to Mosquitto over websockets fails using HTTP/2 due to a bug in Mosquitto. It works fine for chrome and other browsers. The issue in discussion in this thread: https://github.com/eclipse/mosquitto/issues/1211
The workaround is to disable Websocket SPDY in Firefox which in turn prevents Firefox to use HTTP/2. This is of course not a good solution as it disables HTTP/2 for everything.
Until Mosquitto fixes this bug the best option is to build Mosquitto using libwebsockets with http/2 disabled. Note that Mosquitto doesn't benefit from http/2 so disabling it is not a concern. 

Update: As of version 2.0.14, Debian packages from http://repo.mosquitto.org are build with newer libwebsockets.  

## Summary

This plugin manages a mosquitto MQTT message bus broker on behalf of the Hub including authentication and authorization of clients. The MQTT message bus is used by devices to publish Thing Description (TD) and events, and receive actions that are published by plugins and consumers. 

This plugin generates the Mosquitto configuration and launches an instance of the Mosquitto broker. Clients are authenticated on connecting and publications are authorized based on their role.

No manual setup is required other than that mosquitto is installed.


### Authentication

Mosquitto is configured with support for two types of authentication:

1. Client certificate authentication for plugins, for Thing devices, and for administrators
2. Username/password authentication for consumers

Certificate based authentication is very simple. If the client has a valid certificate it can connect to the message bus. This plugin doesn't care how the certificate was issued, just that it is verified by the Certificate Authority. The certificate bundle for CA, hub and plugin client certificate is created on hub startup.

The Hub's 'bin/certs' commandline utility generates the CA, Hub and Plugin certificates and can be used to generate client certificates for consumers. The idprov service automates certificate provisioning for IoT devices using out of band secrets. See [IoT Provisioning](https://github.com/wostzone/idprov-standard) for more information.

Username/password authentication is configured to use the included 'mosqauth' plugin for mosquitto to verify the user login ID and JWT access tokens issued by the authn service on login. See authn for more detail.

### Authorization

Mosquitto is configured to use the mosqauth plugin for authorization with ACLs on topic access. Hub authorization is described in the [Hub authorization document](https://github.com/wostzone/docs/authorization.md) and configures Mosquitto's authorization.

The hub authorization uses group roles. Consumers in a group can access IoT devices in the same group depending on their role in that group. 


## Installation

This plugin produces two binaries. The mosquitto authorization plugin 'mosqauth.so', and the mosquitto manager plugin named 'mosquittomgr'.

See the WoST hub for plugin installation instructions.

### System Requirements

The 'mosquitto' MQTT message broker version 2.0.14 or newer must be installed to avoid websocket problems in firefox.

To build this plugin from source the package libmosquitto-dev must be installed.
