# WoST Gateway

The WoST Gateway is the reference implementation of the gateway for the Web of Secured Things (WoST). It receives information from 'WoST Things' and makes the result available to consumers. The gateway aims to be compatible with the WoT open standard but is constrainted to features that meet the WoST security mandate of "Things Do Not Run Servers".

## Project Status

This project is in the design phase [Jan 2021]. 
Development of the core starts in Feb 2021, followed by plugins.

## Audience

This project is aimed at software developers and system implementors with knowledge of operating systems and computing devices. 

A Binary distribution of the gateway and its plugins can be installed and used by users with basic Linux skills.

## Using The Gateway

The gateway can be used in many different scenarios. The WoT architecture describes [several use-cases](https://www.w3.org/TR/wot-architecture/#sec-use-cases) such as smart home for consumers and smart factories in industry.

There are two typical scenarios. The first is using WoST compliant IoT devices (WoST Things). These devices automatically discovery the gateway on a local network and requests to provision themselves. The administrator logs into the gateway and accepts the provisioning request. From then on the device sends its data to the gateway from where it can be accessed and monitored. A management plugin lets the user administer and monitor devices through a web browser. 

The second scenario works with *legacy* devices that are not WoST compliant. Until manufacturers embrace the security that WoST brings, most devices fall into this category. The gateway has plugins that know how to communicate to these legacy devices. These so-called *protocol bindings* discover and communicate with these devices and push their data into the gateway. 

In both cases the IoT devices do not access the Internet directly. When local, end users connect to the gateway using their web browser to manage and view the 'Thing' information. This works entirely stand-alone and no Internet access is required.
For remote access and monitoring, the gateway can connect securely to a cloud WoST Gateway over the Internet. There is never a direct connection from the Internet to the local network.

This is just a basic example of the possibilities. The gateway works entirely using plugins makes it very flexible and easy to upgrade individual features. Plugin on the gateway can be enabled or disabled as needed.

## Installation

The WoST Gateway is designed to run on Linux based computers. Mac, Windows and Android versions are currently not considered since the recommendation is to run it on a stand-alone computer. 

### System Requirements

Unless the gateway is provided as part of an appliance, it needs to be installed on a computer. It is best to use a dedicated computer for this. 

For home users a raspberry pi 4 will be more than powerful enough to run the gateway. For industrial or automotive usage a dedicated embedded computer system is recommended.

### Install From Package Manager

Installation from package managers is currently not available.

### Install From Binary

Binaries of 64 bit Intel processors will be made available. Create the following folder structure to install the files:

For user based installations:
* /home/{user}/bin/wost/bin      gateway and plugin binaries
* /home/{user}/bin/wost/config   gateway and plugin configuration
* /home/{user}/bin/wost/logs     gateway and plugin logging output

For system installation:
* /opt/wost/                     gateway and plugin binaries
* /etc/wost/                     gateway and plugin configuration
* /var/log/wost/                 gateway and plugin logging output

### Install From Source

To install the core and bundled plugins from source, a Linux system with golang and make tools must be available. 3rd party plugins are out of scope for these instructions and can require nodejs, python and golang.

Prerequisites:
1. Golang 1.14 or newer
2. GCC Make

Build and install from source (tentative):
```
$ git clone https://github.com/wostzone/gateway
$ make build
$ make install
```

After the build is complete, the distribution binaries can be found in the 'dist' folder. 

## Configuration

The gateway is configured through the 'gateway.yaml' configuration file that can be edited with a regular text editor.

~~~yaml 
# gateway.yaml:
host: localhost:9678       // addres of the service bus
protocol: ""               // protocol, eg mqtt, nsq, ...  Default "" is the
plugins:                   // list of plugins to launch 
  wost 
  discovery 
  directory
~~~


The gateway looks in the ./config folder or the /etc/wost folder for this file. This file is optional. Out of the box defaults will provide a working gateway with an internal service bus that listens on localhost port 9678 (WOST). 

Plugins can optionally be configured through yaml configuration files in the same configuration folder.


## Launching

The gateway can be launched manually by invoking the 'wost-gateway' app in the wost folder.

A systemd launcher can be configured to launch automatically on startup for Linux systems that use systemd.

The stg.service file must be copied into the /etc/systemd/system folder after being configured to run as the intended user.

# Design 

![Design Overview](./docs/gateway-design.png)

## Overview

The gateway consists of an internal service bus and a collection of plugins. The plugins fall into two categories, protocol bindings and services. Protocol bindings connect with Things and 3rd party IoT devices while services provide consumer side functionality such as directory services. When available the WoT specified data and API definitions are used.

All features of the gateway are provided through these plugins. Plugins can be written in any language, including ECMAScript to be compliant with the [WoT Scripting API](https://www.w3.org/TR/wot-architecture/#sec-scripting-api). It is even possible to write a plugin for plugins to support a particular programming platform such as EC6. 

As mentioned, plugins fall into two categories depending on their purpose:
* Protocol bindings provide connectivity for WoST Things and for 3rd party protocols. These plugins convert the device description data they receive to a Thing Description document and submit events in the WoT format according to the WoT specifications.
* Service plugins typically subscribe to Thing updates to provide a service to client applications. They can also publish actions for Things to execute. Services can make additional API's available to consumers, for example a directory service and a web client interface.

The internal service bus provides the messaging infrastructure for communication between plugins. It can be implemented by an existing message bus or queuing service. By default a built-in lightweight internal message queue service is used. For high performance use-cases it can be replaced with a beefier service such as NSQ, MQTT, and AMQP. The included gateway client library implements support for the various message queuing services and hides the message bus implementation from the plugins. This allows the re-use of a plugin in places with different message bus implementations. The gateway client library will be made available in various programming language such as golang, ES6, and Python. 

While all plugins are optional, a few are neccesary for normal operation. 

By default the internal service bus is started by the gateway and listens on localhost. Only local plugins are able to connect. 

Plugins publish and subscribe to data channels. WoT related channels, such as a TD update channel, events, and actions are predefined. Plugins can add additional channels as needed.


## Protocol Binding Plugins

The primary role of protocol binding plugins is to translate to and from the WoT data formats to publish TDs, events and receive actions. This involves the respective 3 channels 'td', 'event', and 'action'.

The format of the data pushed into the channel MUST match the schema associated with the channel ID. Schemas are defined in the JSON-LD format as defined in schema.org, WoT schemas, and NGSI-LD schemas. 

For example, the schema for the [Thing Description](https://www.w3.org/TR/wot-thing-description/#behavior-data) is descripted in the [TD Schema](https://www.w3.org/TR/wot-thing-description/#json-schema-for-validation)


## Service Plugins

Service plugins consume and optionally convert channel data. They can run their own web server to make this data available to consumers. For example a directory service provides an API to query known devices.

Service plugins can optionally publish transformed data onto the TD/event channels or creat a whole new channel specific to the purpose of the plugin and associated consumer plugins.

<todo>See the list of available service plugins for details. </todo>

## Writing Plugins

Plugins can be written in any programming language. They can include a configuration file that describes their purpose and the pipeline they use. Plugins must use the gateway library to connect to the service bus.

There is nearly no boilerplate code involved in writing plugins, except for adhering to the channel data requirements. Plugins can therefore be very lightweight and efficient. 

Plugins run in their own process, isolated from other plugins. It is however possible to write a plugin that launches other plugins in threads. For example, a JS plugin can load additional plugins written in Javascript. Each of the additional plugin connects to the gateway channels using the client library.

### Data Channels

Data published on WoST Gateway channels MUST adhere to that data channel's schema specification.  The Gateway has the following predefined channels:
* td: Thing descriptions https://www.w3.org/TR/2020/WD-wot-thing-description11-20201124/
* event: Thing events
* action: Thing actions.

Message published on these channel MUST adhere to the WoT data schemas for TD, events and actions.

To publish on a data channel simply connect to the channel address:
> https://{host}/{channelID} and publish the message.
To receive channel data, connect to the channel and include an optional filter:
> https://{host}/{channelID}?device=id, to listen to a specific device/


### Reference Plugin s
The WoST project plans to include several plugins for working out of the box.

* The 'discovery' protocol binding announces the gateway on the local network using mDNS. This is intended to let Secured Things discover the gateway.

* The  'wost' protocol binding provides a websocket API for use by WoST Things to provision, publish TD's, publish events, and receive actions.

* The 'directory' service plugin provides an HTTPS API for consumers to query provisioned and discovered Things. 

* The 'intermediary' service plugin forwards TD and events from exposed Things to a remote gateway or intermediary, and optionally receive actions. This is intended for cloud based access to Things.

* The 'history' service plugin implements the history API to query historical values for Things. (NGSI-LD history API)

* The 'script' service plugin executes ECMA scripts. It lets script receive channel data and can execute actions. 

* The 'notification' service plugin sends messages from the notification channel to the configured destination, eg Email, SMS, other.

* The 'swui' service plugin provides a simple web based interface to view and manage Things. It supports the history and script API.


All of these plugins can be substituted by another implementation as needed. 


## Launching Plugins

Plugins are launched at startup and given three arguments: 
* {host} containing the IP and port of the service bus connection.
* {authorization} containing the authorization token the plugin must include when establishing its connection.
* {configFile} containing the path to the plugin YAML configuration file. This file is optional. If possible plugins should function out of the box without configuration.

## Service Bus Connection

After launch, plugins connect to their channels. The default address for the internal service bus is made up as follows:
> ws://{host}/{channel}[/{thingID}]

Where:
* {host} is the parameter passed on startup. This depends on the chosen service bus.
* {channel} is the ID of the channel to publish or subscribe. Each channel has an associated schema that describes the data format that is published on the channel. 
* {thingID} is an optional filter to address a single Thing
 
A valid authorization header token must be present. This token is provided on startup. The core will reject any connection requests that do not contain a valid token.

While a plugin can make as many connections as needed it is strongly recommended to adhere to the single responsibility principle and only connect to the channel and stage that is needed to fulfil that responsibility. 

Note that the client libraries implement the connection logic for the various protocols so the plugin developer only need to know the channel ID and optionally the device ID.


# Contributing

Contributions to WoST projects are very welcome. There are many areas where help is needed, especially with documentation and building plugins for IoT and other devices. See [CONTRIBUTING](docs/CONTRIBUTING.md) for guidelines.


# Credits

This project builds on the Web of Things (WoT) standardization by the W3C.org standards organization. For more information https://www.w3.org/WoT/

