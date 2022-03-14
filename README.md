# WoST Hub

The WoST Hub is the reference implementation of the Hub for the *Web of Secure Things*. It acts as an intermediary between IoT devices 'Things' and consumers. Consumers interact with Things through Hub services without connecting directly to the Thing device. 

## Project Status

Status: The status of this plugin is Alpha. It is functional but breaking changes can be expected.


## Audience

This project is aimed at software developers and system implementors that share concerns about the security and privacy risk of IoT devices. WoST users choose to not run servers on Things and instead use a hub and spokes model.

## Objective

The objective of WoST is to support the internet of things in a highly secure manner. The core mandate is that 'Things Do Not Run Servers'. The WoST Hub supports this objective by simplifying IoT device implementation and by isolating IoT devices from possible attacks via a secure Hub. 

The intent is to follow the WoT and other open standard where possible.


## Summary
This document describes a technical overview of the WoST Hub. A [user manual](user-manual.md) is under development.

Security is big concern with today's IoT devices. The Internet of Things contains billions of devices that when not properly secured can be hacked. Unfortunately the reality is that the security of many of these devices leaves a lot to be desired and are never upgraded with security patches. This problem is only going to get worse as more IoT devices are coming to market. Imagine a botnet of a billion devices on the Internet ready for use by in-scrupulous actors. 

WoST compatible devices are much more secure simply by not running a server and by not being directly accessible. Instead, WoST devices only connect to the Hub that is provisioned. This removes a large attack vector and greatly reduces the need for security updates on these devices.

This 'WoST Hub' repository provides core services to securely interact with IoT devices and consumers. This includes certificate management, authentication, authorization, provisioning, message bus service and directory service.

WoST compatible IoT devices therefore do not need to implement these features. This not only reduces required device resources such as memory and CPU (and cost) but also improves security by isolating the IoT device from its consumers with less need to 'harden' devices for security. An additional benefit is that consumers receive a consistent user experience independent of the IoT device provider as all interaction takes place via the  Hub interface. 

WoST is based on the 'WoT' (Web of Things) open standard developed by the W3C organization. It aims to be compatible with this standard.


### Plugins

All Hub functionality is provided through plugins and can be extended with additional plugins. Plugins can be protocol bindings to bridge different IoT technologies, or services to enrich the IoT data that has been collected. All core services are written as plugins and can be replaced if desired.

Plugins can be written in any programming language but must follow some simple guidelines. The [writing-plugins.md] document describes how to write new plugins. Existing services/plugins can also serve as an example.  


## Core Services

### Ports

The Hub includes several services that listen on specific ports. The default port numbers used by core hub services are:

* 8880 idprov provisioning for discovery of hub by IoT devices
* 8881 wost hub authentication service for token creation and renewal
* 8882 wost bridge service for linking two hubs
* 8883 mqtt message bus, mqtt port requiring username-password authentication
* 8884 mqtt message bus, mqtt port requiring certificate authentication
* 8885 mqtt message bus, websocket port requiring username-password authentication
* 8886 thingdir thing directory service port for querying known Thing Description documents
* 8443 thingview web based thing viewer application for managing and viewing things
* 8443 [Tentative] A proxy service to provide a single HTTPS API access point for the services.  


### Launcher Service

The Hub launcher is responsible for starting and stopping other Hub services. Its purpose is to launch services and monitor their status.

### certs: Certificate Management

The certs service provides a commandline interface for managing certificates.
- the Hub self-signed CA certificate. Can be added to the browser for local use.
- the Hub server certificate, signed by the CA. 
- the Hub plugin client certificate, signed by the CA. Intended for machine-to-machine authentication.
- the IoT device certificates, used by the 'idprov' service during the provisioning process. 

### idprov: Provisioning Service

IoT devices that support the [idprov protocol](https://github.com/wostzone/idprov-standard) can automatically discover the provisioning server on the local network using the DNS-SD protocol and initiate the provisioning process. When accepted, a CA signed certificate is issued. This certificate supports machine to machine authentication between IoT device and Hub Services such as the message bus along with a list of services that can be used to connect to. See [idprov service](https://github.com/wostzone/hub/tree/main/idprov) for more information. 

In order to use provisioning IoT devices must support HTTPS/TLS clients with SSL encryption.

### authn: Authentication Service

The authentication service manages users and issues access and refresh tokens.
It provides a CLI to add/remove users and a service with a REST API to handle authentication request and issue tokens. See [authn service](https://github.com/wostzone/hub/tree/main/authn) for more information.

### authz: Authorization Service

The authorization service manages role based access control using groups of Consumers and Things.
Consumers that are in the same group as a Thing have permission to access the Thing based on their role as viewer, operator, manager, administrator or thing. See the [authorization service](https://github.com/wostzone/hub/tree/main/authz) for more information. 

### mosquittomgr: Message Bus Manager and Mosquitto auth plugin

Interaction with Things takes place via the MQTT message bus. Things publish their TD document and events onto the bus and subscribe to action messages. Services and consumers can subscribe to these publications and publish messages to the Thing. This is the only method to communicate with the Thing. 

The Mosquitto manager configures the Mosquitto MQTT broker (server) including authentication and authorization of things, services and consumers. See the [mosquittomgr service](https://github.com/wostzone/hub/tree/main/mosquittomgr) for more information.

IoT devices must be able to connect to the MQTT message bus through an MQTT client with encryption and certificate authentication. The [Eclipse Paho](https://www.eclipse.org/paho/) project provides client libraries for many different programming environments. 

### thingdir: Directory Service 

The directory service provides a REST API for consumers to list or query known Things. The service stores the TD Documents published by Things. It uses the Authorization service to filter the TD's that a consumer is allowed to see. See the [directory service](https://github.com/wostzone/hub/tree/main/thingdir) for more information.

## Installation (draft)

The WoST Hub is designed to run on Linux based computers. It might be able to work on other platforms but at this stage this is not tested nor a priority.

### System Requirements

It is recommended to use a dedicated server or container for installing the Hub and its plugins. For experimental industrial or automotive usage a dedicated embedded computer system with Intel or ARM processors is recommended. For home users a raspberry pi 2/3/4+ will be sufficient to run the Hub and most plugins.

The minimal requirement is 100MB of RAM and an Intel Celeron, or ARMv7 CPU. Additional resources might be required for some plugins. See plugin documentation.

* mosquitto:
The Hub requires the installation of the Mosquitto MQTT message broker version 2.0.14 or newer. To build from source the libmosquitto-dev package must be installed as well.
* The 'mosquittomgr' plugin manages the configuration and security of the Mosquitto broker on behalf of the Hub. Other MQTT brokers can be used instead of Mosquitto but will require an accompanying service to handle authentication and authorization. The MQTT broker can but does not have to run on the same system as the Hub.

### Install From Package Manager

Installation from package managers is currently not available.

### Install From Binary Releases

Beta and production releases will include binaries for amd64 and arm64 (pi 2-4).

When installing manually using binaries the manual configuration process must be followed as described in installation from source.  

### Manual Install As User

The Hub can be installed and run as a dedicated user or system user. This section describes to install the Hub in a dedicated user home directory. 

1. Create a user, for example a 'wosthub' user. Login as that user.
2. Create the hub folder structure:

```sh
mkdir -p ~/bin/wosthub/bin
mkdir -p ~/bin/wosthub/config
mkdir -p ~/bin/wosthub/logs 
mkdir -p ~/bin/wosthub/certs 
mkdir -p ~/bin/wosthub/certstore
```

3. Copy the application binaries into the bin folder and default configuration in the config folder
```sh
cp bin/* ~/bin/wosthub/bin
cp config/* ~/bin/wosthub/config
```

4. Generate the certificates using the certs CLI

```sh
cd ~/bin/wosthub
bin/certs certbundle   
```

5. Install Mosquitto v2.0.14+ using the package manager of choice

For example on Ubuntu:
> sudo apt install mosquitto

Note 1: Do not autostart or configure mosquitto. Its default configuration will not be used. The 'mosquittomgr' service will launch a Mosquitto instance with a generated configuration and dedicated authentication/authorization plugin.

Note 2: As of Early 2022, Mosquitto v2.0.14 or newer is required as it is built with libwebsockets that has http2 disabled. Mosquitto doesn't handle http/2 properly and results in Firefox being unable to connect.



### Manual Install To System (tenative)

For systemd installation to run as user 'wosthub'. When changing the user and folders make sure to edit the init/wosthub.service file accordingly. From the dist folder run:

1. Create the folders and install the files

```sh
sudo mkdir /opt/wosthub/       
sudo mkdir -P /etc/wosthub/certs/ 
sudo mkdir -P /var/lib/wosthub/certstore/ 
sudo mkdir /var/log/wosthub/   

# Install WoST configuration and systemd
# download and extract the binaries tarfile in a temp for and copy the files:
tar -xf wosthub.tgz
sudo cp config/* /etc/wosthub
sudo vi /etc/wosthub/hub.yaml    - and edit the config, log, plugin folders
sudo cp init/wosthub.service /etc/systemd/system
sudo cp bin/* /opt/wosthub
```

2. Setup the system user and permissions

```sh
sudo adduser --system --no-create-home --home /opt/wosthub --shell /usr/sbin/nologin --group wosthub
sudo chown -R wosthub:wosthub /etc/wosthub
sudo chown -R wosthub:wosthub /var/log/wosthub
sudo chown -R wosthub:wosthub /var/lib/wosthub

sudo systemctl daemon-reload
```

3. Install mosquitto v2.0.14+ on Ubuntu but do not configure it:


```sh
sudo apt install mosquitto
```

4. Start the hub

```sh
sudo service wosthub start
```

5. Autostart the hub after startup

```sh
sudo systemctl enable wosthub
```


### Build From Source

To build the core and bundled plugins from source, a Linux system with golang 1.17+ and make tools must be available on the target system. 3rd party plugins are out of scope for these instructions and can require nodejs, python and golang.

Prerequisites:

1. Golang 1.17 or newer 
2. GCC Make

Build from source (tentative):

```sh
$ git clone https://github.com/wostzone/hub
$ cd hub
$ make all
```

After the build is complete, the distribution binaries can be found in the 'dist/bin' folder and configuration files in dist/config.

To install the hub as the user:

```sh
make install
```

This copies the binaries and config to the ~/bin/wosthub location as described in the manual install section. Executables are always replaced but only new configuration files are installed. Existing configuration remains untouched.  

Additional plugins are built similarly:

```bash
$ git clone https://github.com/wostzone/{plugin}
$ cd {plugin}
$ make all 
$ make install                    (to install as user to ~/bin/wosthub/...)
```

## Configuration

All Hub services will run out of the box with their default configuration. To change the default network and folder locations edit the 'config/hub.yaml' configuration file. 

Hub services load their common configuration from the hub.yaml file in the config folder. This file MUST exist as it contains the message bus connection information for use by plugins. If no address is configured, the host outbound IP address is determined during startup. For hosts with multiple addresses, the address to use can be configured in hub.yaml

Plugins can have their own plugin specific configuration file in the config folder. Plugins must be able to run without a configuration file.

## Launching

The Hub can be launched manually by invoking the 'launcher' app in the wost bin folder. eg ~/bin/wosthub/bin/launcher. The services to start are defined in the config/launcher.yaml configuration file. When adding services, this file needs to be updated with the new service executable name.

A systemd launcher is provided that can be configured to launch on startup for systemd compatible Linux systems. See 'init/wosthub.service'

```sh
sudo cp init/wosthub.service /etc/systemd/system
sudo vi /etc/systmd/system/wosthub.service      (edit user and working directory)
sudo systemctl enable wosthub
sudo systemctl start wosthub
```

## Plugin Installation 

Additional plugins are installed in the wosthub 'bin' directory. It is also possible to create a softlink from this directory to location of the actual binary.

After downloading or building the plugin executable:

1. Copy the plugin binary into the Hub binary folder, eg ~/bin/wosthub/bin or /opt/wosthub.
2. Copy the plugin configuration file {plugin}.yaml to the Hub configuration folder, eg ~/bin/wosthub/config or /etc/wosthub.
3. Add the plugin to the launcher.yaml configuration file in the configuration folder. The new plugin will be started automatically when the hub starts. Note that plugins start in the listed order.

# Design

![Design Overview](./docs/hub-overview.png)

## Overview

The Hub operates using a message bus, a curated set of core plugins and optional additional plugins. The plugins fall into two categories, protocol bindings and services:

* Protocol bindings provide connectivity for WoST Things and legacy/3rd party IoT devices. For example, the idprov protocol binding provides IoT devices with the ability to self-provision. Legacy protocol bindings convert the legacy device description to a WoT compliant Thing Description (TD) document and submit these onto the Hub message bus. Actions received from the message bus are passed back to the device after converting it into the device's native format.

* Services provide a service to consumer applications. They can receive requests and publish actions for Things to execute. Services can make additional API's available to consumers, for example the directory service provides an API to query for Things. Communication from consumers to Things goes via Hub services. Consumer applications can also access the message bus directly to exchange messages with things but will require proper authentication using the authn service.

## Hub Message Bus

Central to the Hub is its MQTT publish/subscribe message bus. Messages sent over this message bus are WoT compatible and conform to the format defined in the [WoT TD standard](https://www.w3.org/TR/wot-thing-description/). 

WoST uses Mosquitto as the MQTT message bus, which is configured and managed using the core 'mosquittomgr' service. 'mosquittomgr' handles the configuration, authentication and authorization of connections on the message bus. 

Other MQTT implementations can be used instead but will require their own manager service to handle configuration, authentication and authorization.

Separate ports are used to support authentication using client certificates, username/password authentication using the MQTT protocol, and username/password authentication over Websockets. The ports are defined in the hub.yaml configuration file.

## Protocol Binding Plugins

Protocol Binding plugins adapt 3rd party IoT protocols to WoT TD publications on the MQTT message bus. This turns the 3rd party devices into WoST compatible Things. For example, an openzwave protocol binding makes ZWave devices available via the Hub.

Consumers are agnostic to the IoT device protocols used and only need to access the WoST services. Protocol Bindings (and IoT devices) also do not need knowledge of authentication and authorization as this is handled via the message bus plugin.  

## Service Plugins

Service plugins provide their own API to the consumer. For example the core directory service plugin provides the Directory API to query for Things.

Service plugins subscribe to TD and Event messages on the message bus to obtain information about things and can publish actions to control Things.

Service plugins included with the hub are the 'authn' authentication service, the 'thingdir' directory service, and the 'thingview' web client. Plugins for additional functionality should be installed as needed.   

## Writing Plugins

Plugins can be written in any programming language. The Hub provides a client library (hub/lib/client) in Golang to easily connect to the message bus. This library will be maintained along with the hub. Implementations in Python and Javascript are planned for the future.

Plugins run in their own process, isolated from other plugins.

A [scripting API](https://www.w3.org/TR/wot-scripting-api/) plugin is also planned and can be used to create plugins. (todo)

See [the documentation on writing plugins](docs/writing-plugins.md).

## Launching Plugins

Core plugins are launched at startup by the Hub launcher and accept the Hub arguments to determine configuration files and folders. See 'launcher --help' for details. The default settings work out of the box.

Most plugins have an optional configuration file named {pluginID}.yaml in the {wosthub}/config folder. A default file is provided with the plugin that describes the available options.

## Client Library

The WoST project provides a [WoST client library](https://github.com/wostzone/hub/lib/client). This library provides functions for building WoST IoT clients and plugins including connections to the MQTT message bus and to construct WoT compliant Thing Description models (TD).

IoT devices will likely also use the [provisioning client](https://github.com/wostzone/hub/idprov/pkg/idprov) to automatically discovery the provisioning server and obtain a certificate used to connect to the message bus.

The above library is written in golang Python and Javascript Hub API libraries are planned. They will be added to https://github.com/wostzone/lib/{python}|{js}|{...}

# Contributing

Contributions to WoST projects are always welcome. There are many areas where help is needed, especially with documentation and building plugins for IoT and other devices. See [CONTRIBUTING](docs/CONTRIBUTING.md) for guidelines.

# Credits

This project builds on the Web of Things (WoT) standardization by the W3C.org standards organization. For more information https://www.w3.org/WoT/

This project is inspired by the Mozilla Thing draft API [published here](https://iot.mozilla.org/wot/#web-thing-description). However, the Mozilla API is intended to be implemented by Things and is not intended for Things to register themselves. The WoST Hub will therefore deviate where necessary.

Many thanks go to JetBrains for sponsoring the WoST open source project with development tools.  
