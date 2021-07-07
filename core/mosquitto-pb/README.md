# WoST MQTT Mosquitto Binding

This plugin manages Mosquitto, a lightweight and powerful MQTT message bus broker. In order to use this plugin the mosquitto broker must be installed.

MQTT is the primary communication method for the Hub and its plugins. Things and consumers can connect via MQTT to publish and subscribe to TD's, Events and Actions as described in the protocol binding. 

## Project Status

Status: early alpha. 

What it does:
- configure and launch a Mosquitto instance using the hub configuration
- configure mosquitto logging
- configure concertificate based authentication for devices, plugins and administrators
- configure external authorization as per config file

## Audience

This project is aimed at web-of-things developers that share concerns about the security and privacy risk of running a server on every WoT Thing. WoST developers choose to not run servers on Things and instead use a hub and spokes model. The WoST project provides this Hub.

## Summary

This plugin manages a mosquitto MQTT message bus broker on behalf of the Hub including authentication and authorization of clients. The MQTT message bus is used by devices to publish Thing Description (TD) and events, and receive actions that are published by plugins and consumers. 

This plugin generates the Mosquitto configuration and launches an instance of the Mosquitto broker. Clients are authenticated on connecting and publications are authorized based on their role.

No manual setup is required other than that mosquitto is installed.

### Topic Structure

The MQTT topic structure is as follows:
>  things/{publisherID}/{thingID}/td|event|action

{publisherID} is the ID of the publishing device. In case of plugins it is the plugin instance ID. In case of IoT devices it is the ThingID of the IoT Device. The publisher ID is used to ensure a unique topic for all Things.


### Authentication

Mosquitto is configured with certificate based authentication for plugins, for Thing devices, and for administrators. Other consumers authenticate with a login ID and password.

Certificate based authentication is very simple. If the client has a valid certificate it can connect to the message bus. Without it, it needs a login ID and password. 

This plugin doesn't care how the certificate was issued, just that it is verified by the Certificate Authority. See [IoT Provisioning](https://github.com/wostzone/idprov-standard) on how devices can obtain a certificate. The plugin certificate is created on hub startup and available to plugins only.

The administrator issues username and password to consumers. The administrator uses the admin interface to this plugin to add and remove users. This plugin also includes a CLI (commandline interface) to administer users.

The Hub's 'cmd/gencert' package generates a commandline utility to generate the CA, Hub and Plugin certificates.

### Authorization

Mosquitto is configured to use a plugin for authorization with ACL on topic access. The actual authorization is handled by the authorization plugin. It authorizes access to Thing topics based on a client's role in the same group as the Thing.

Hub authorization is described in the [Hub authorization document](https://github.com/wostzone/docs/authorization.md) and is independent from Mosquitto's concept of authorization.

The 'mosqplug' sub package installs a mosquitto plugin to integrate with the Hub's authorization.

## Installation

This protocol binding produces two binaries. The mosquitto authorization plugin 'mosqauth.so', and the mosquitto protocol binding named 'mosquitto-pb'.

See the WoST hub for plugin installation instructions.
'make install' install the plugin in the ~/bin/wost/bin folder.

This plugin is started by the hub if the hub.conf file includes it in its plugins section.

### System Requirements

The 'mosquitto' MQTT message broker version 1.5 or newer must be installed.

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
