# Mosquitto Manager 

This plugin manages the Mosquitto configuration, a lightweight and powerful MQTT message bus broker. In order to use this plugin the mosquitto broker must be installed.

MQTT is the primary communication method for the Hub and its plugins. Things and consumers can connect via MQTT to publish and subscribe to TD's, Events and Actions as described in the protocol binding. 

## Project Status

Status: early alpha. 

What it does:
- configure and launch a Mosquitto instance using the hub configuration
- configure mosquitto logging
- configure authentication using client certificates 
- configure authentication using the included username/password authentication plugin
- configure group/role authorization using the included ACL plugin

### Known Issues

Firefox connections to Mosquitto over websockets fails using HTTP/2 due to a bug in Mosquitto. It works fine for chrome and other browsers. The issue in discussion in this thread: https://github.com/eclipse/mosquitto/issues/1211
The workaround is to disable Websocket SPDY in Firefox which in turn prevents Firefox to use HTTP/2. This is of course not a good solution as it disables HTTP/2 for everything.
Until Mosquitto fixes this bug the best option is to build Mosquitto using libwebsockets with http/2 disabled. Note that Mosquitto doesn't benefit from http/2 so disabling it is not a concern. 

Update: As of version 2.0.14, Debian packages from http://repo.mosquitto.org are build with newer libwebsockets.  

## Audience

This project is aimed at web-of-things developers that share concerns about the security and privacy risk of running a server on every WoT Thing. WoST developers choose to not run servers on Things and instead use a hub and spokes model. The WoST project provides this Hub.

## Summary

This plugin manages a mosquitto MQTT message bus broker on behalf of the Hub including authentication and authorization of clients. The MQTT message bus is used by devices to publish Thing Description (TD) and events, and receive actions that are published by plugins and consumers. 

This plugin generates the Mosquitto configuration and launches an instance of the Mosquitto broker. Clients are authenticated on connecting and publications are authorized based on their role.

No manual setup is required other than that mosquitto is installed.


### Authentication

Mosquitto is configured with support for two types of authentication:

1. Client certificate authentication for plugins, for Thing devices, and for administrators
2. Username/password authentication for consumers

Certificate based authentication is very simple. If the client has a valid certificate it can connect to the message bus.
This plugin doesn't care how the certificate was issued, just that it is verified by the Certificate Authority. The certificate bundle for CA, hub and plugin client certificate is created on hub startup.

The Hub's 'cmd/gencert' commandline utility generates the CA, Hub and Plugin certificates and can be used to generate client certificates for consumers. The idprov service automates certificate generation for IoT devices using out of band secrets. See [IoT Provisioning](https://github.com/wostzone/idprov-standard) for more information.

Username/password authentication is configured to use the included 'mosqauth' plugin for mosquitto. This plugin uses the hub's auth package for verify the password with the stored hash. The password store can be updated with the 'auth' commandline utility. The administrator issues username and password to consumers. Currently the administrator will have to use the 'auth' utility to manage passwords. A password service will be added in the future to support a web interface for administrators and users to change their own password. The password will also be valid for other supporting services such as the client dashboard.


### Authorization

Mosquitto is configured to use the mosqauth plugin for authorization with ACLs on topic access. Hub authorization is described in the [Hub authorization document](https://github.com/wostzone/docs/authorization.md) and configures Mosquitto's authorization.

The hub authorization uses group roles. Consumers in a group can access IoT devices in the same group depending on their role in that group. 


## Installation

This plugin produces two binaries. The mosquitto authorization plugin 'mosqauth.so', and the mosquitto manager plugin named 'mosquittomgr'.

See the WoST hub for plugin installation instructions.
'make install' install the plugin binary in the ~/bin/wost/bin folder and the configuration files in ~/bin/wost/config.

This plugin is started by the hub. It requires that the plugin is included in the wost.yaml configuration which is the default.

### System Requirements

The 'mosquitto' MQTT message broker version 2.0.14 or newer must be installed to avoid websocket problems in firefox.

To build this plugin from source the package libmosquitto-dev must be installed.

### Build From Source

Build and install from source (tentative):
```
$ git clone https://github.com/wostzone/mosquitto-pb
$ make all 
```
The plugin can be found in dist/bin for 64bit intel or amd processors, or dist/arm for 64 bit ARM processors. Copy this to the hub bin or arm directory.

An example configuration file is provided in config/mosquitti-pb.yaml, as is a template for mosquitto configuration. Copy these to the hub config directory.

## Credits

This protocol binding is inspired by the Mozilla Thing draft API [published here](https://iot.mozilla.org/wot/#web-thing-description). However, the Mozilla API is intended to be implemented by Things and is not intended for Things to register themselves. This protocol binding will therefore deviate where neccesary. 

Credit also goes to @hidaris for the willingness to discuss the standardization of the MQTT protocol for use with the Web of Things.
