# Hive-Of-Things Hub

The Hub for the *Hive of Things* is an intermediary between IoT devices 'Things', services, and consumers using a hub-and-spokes architecture. Consumers interact with Things via the Hub without connecting directly to the Thing device. The Hub uses the [cap'n proto](https://capnproto.org/) for Capabilities based secured communication.

## Project Status

Status: The status of the Hub is In Development. It is undergoing a rewrite to use **capnp** for infrastructure.

HiveOT is based on the [W3C WoT TD 1.1 specification](https://www.w3.org/TR/wot-thing-description11/). See [docs/README-TD] for more information.

## Audience

This project is aimed at software developers and system implementors that are working on secure IoT devices. Users choose to not run servers on Things and instead use a hub and spokes model which greatly reduces the security risk post by traditional IoT devices.

## Objective

The primary objective of HiveOT is to provide a solution to use 'internet of things' devices in a highly secure manner. The Hub supports this objective by not allowing IoT devices to receive incoming network connections and by isolating IoT devices from the wider network via a secure Hub. Instead, IoT devices provide Capability services and connect to the Hub. By leveraging Capabilities based design, security is greatly improved while allowing for a largely decentralized approach to services.

> The HiveOT mandate is: 'Things Do Not Run (TCP) Servers'.

The secondary objective is to simplify development of IoT devices for the web of things. HiveOT supports this by requiring only minimal features to operate on an IoT device. Complexity and resource usage is kept minimal by not running a web server.    The Hub handles authentication and authorization on behalf of the device. 

The third objective is to follow the WoT and other open standard where possible.

## Summary

This document describes a technical high level overview of the Hub.

Security is big concern with today's IoT devices. The Internet of Things contains billions of devices that when not properly secured can be hacked. Unfortunately the reality is that the security of many of these devices leaves a lot to be desired. Many devices are vulnerable to attacks and are never upgraded with security patches. This problem is only going to get worse as more IoT devices are coming to market. Imagine a botnet of a billion devices on the Internet ready for use by unscrupulous
actors.

This 'HiveOT Hub' repository provides core services to securely interact with IoT devices and consumers. This includes certificate management, authentication, authorization, provisioning, message bus service and directory service.

HiveOT compatible IoT devices (Things) therefore do not need to implement these features. This improves security as IoT devices do not run Web servers and are not directly accessible. They can remain isolated from the wider network and only require an outgoing connection to the Hub. This in turn reduces required device resources such as memory and CPU (and cost). An additional benefit is that consumers receive a consistent user experience independent of the IoT device provider as all
interaction takes place via the Hub interface.

HiveOT is based on the 'WoT' (Web of Things) open standard developed by the W3C organization. It aims to be compatible with this standard.

The communication infrastructure for the services is provided by 'Cap'n Proto', or capnp for short. Capnp provides a Capabilities based RPC for service invocation that is inherently secure. Only clients which have obtained a valid 'Capability' can invoke that capability, eg read a sensor or control a switch. The RPC will only pass requests that are valid, so the device does not have to concern itself with authentication and authorization. 

Since the Hub acts as the intermediary, it is responsible for features such as authentication, logging, resiliency, pub/sub and other protocol integration. The Hub can dynamically delegate some of these services to devices that are capable of doing so, potentially creating a decentralized solution that can scale as needed and recover from device failure. As a minimum the Hub manages service discovery acts as a proxy for capabilities. 

## Services

All Hub features are provided through services. The first service is the Hub gateway which is the entry point into the Hub. When a clients connects to the gateway it  receives a set of 'capabilities' for authentication and other available features.

Some of the services that are planned for the Hub:
- Hub Gateway to which all clients (devices, consumers and other services) connect. The Hub gateway provides other service capabilities depending on the client. Multiple instances of the Hub gateway can exist on the network to offer redundancy and failover.   
- Directory Service provides the capabilities to register Things by IoT devices and to query for available Things by consumers.
- State Service offers services or clients an easy way to persist state into a configured backend.
- Pub/Sub service offers other services an easy way to publish and subscribe messages onto a message bus. This is intended for notifying of events, actions and TD documents.
- Authentication Service provides the capability to issue identity tokens to consumers. Identity tokens are used to determine what IoT Things the consumer can use.  
- Group Service provides the capability to manage the IoT things a consumer has access to.
- Provisioning Service provides the capability to register IoT devices and offer a client certificate. IoT devices use certificates instead of identity tokens to authenticate themselves.
- Logging Service provides the capability to capture logging information and to send it to an external logging service such as Zipkin or other.
- Tracing Service provides the capability to trace requests and measure performance.
- Rate limiting Service limits the maximum rate at which clients can make RPC calls.
- Resiliency Service handles RPC call failures and possibly provide a different service to handle the call.

A second category of services are IoT Protocol Binding services. These services connect with IoT devices using 3rd party protocols such as ZWave, CoAP, Zigbee, LoRaWAN, 1-wire, and others.

IoT devices can also connect to the Hub directly by using the Hub services for IoT devices for provisioning, directory and pub/sub.

Consumers can receive events and control devices via the Hub. The Hub proxies the messages so the consumer only connects to the Hub gateway.

Services can be written in any programming language but must provide and use Hub capabilities using capnp. The [writing-services.md] document describes how to write new services. Existing services/plugins can also serve as an example. Client libraries are available for using capnp in the most popular programming languages.

Hub API's are defined with the [capnproto schema language](https://capnproto.org/language.html). This is like protobuf but then on steroids. 


### Inter-service communication

Services and compatible clients communicate using capnp RPC. Each service defines its interface and data types in a .capnp file. The compiler generates a corresponding language file in the programming language of choice.  The HiveOT project offers some client libraries to make using these APIs easier and reduce boilerplate.  

Capnp is unique in that it is based on 'capabilities'. In order to invoke an RPC method a client must first have this method capability. This can be offered to the client by another service. This intrinsic security feature is core to the Hub security.  

Capnp supports bi-directional streams, making it possible to publish and subscribe of events, actions and TD documents. 


### idprov: Provisioning Service

IoT devices that support the [idprov protocol](https://github.com/hiveot/idprov-standard) can automatically discover the Hub on the local network using the DNS-SD protocol and initiate the provisioning process. When accepted, a CA signed device (client) certificate is issued.

The device certificate supports machine to machine authentication between IoT device and Hub. See [idprov service](https://github.com/hiveot/hub/tree/main/idprov) for more information.

### authn: Authentication Service

The authentication service manages users and issues access and refresh tokens.
It provides a CLI to add/remove users and a service to handle authentication request and issue tokens. See [authn service](https://github.com/hiveot/hub/tree/main/authn) for more information.


### groups: Manages groups of users and Things 

The group service manages groups that contain consumers and Things.
Consumers that are in the same group as a Thing have permission to access the Thing based on their role as viewer, operator, manager, administrator or thing. See the [authorization service](https://github.com/hiveot/hub/tree/main/authz) for more information.

### mosquittomgr: Message Bus Manager and Mosquitto auth plugin (deprecated)

This mosquittomgr service is replaced by pub/sub.

Interaction with Things takes place via a message bus. [Exposed Things](https://www.w3.org/TR/wot-architecture/#exposed-thing-and-consumed-thing-abstractions) publish their TD document and events onto the bus and subscribe to action messages. Consumers can subscribe to these messages and publish actions to the Thing.

The Mosquitto manager configures the Mosquitto MQTT broker (server) including authentication and authorization of things, services and consumers. See the [mosquittomgr service](https://github.com/hiveot/hub/tree/main/mosquittomgr) for more information.

IoT devices must be able to connect to the message bus through TLS and use client certificate authentication. The Hub library provides protocol bindings to accomplish this.

### thingdir: Directory Service

The directory service captures TD document publications and lets consumer list and query for known Things. It uses the Authorization service to filter the TD's that a consumer is allowed to see. See the [directory service](https://github.com/hiveot/hub/tree/main/thingdir) for more information.

The directory service is intended for use by consumers. IoT devices only need to use the pub/sub API to publish TDs and events, and subscribe to actions.

## Build

This step can be skipped if you are using the pre-built binaries.

### Build From Source

To build the core and bundled plugins from source, a Linux system with golang 1.15+ and make tools must be available on the target system. 3rd party plugins are out of scope for these instructions and can require nodejs, python and golang.

Prerequisites:

1. Golang 1.17 or newer
2. GCC Make

Build from source (tentative):

```sh
$ git clone https://github.com/hiveot/hub
$ cd hub
$ make all
```

After the build is complete, the distribution binaries can be found in the 'dist/bin' folder and configuration files in dist/config.

The makefile also support a quick install for the current user:

```sh
make install
```

This copies the binaries and config to the ~/bin/hivehub location as described in the manual install section below. Executables are always replaced but only new configuration files are installed. Existing configuration remains untouched.

Additional plugins are built similarly:

```bash
$ git clone https://github.com/hiveot/{plugin}
$ cd {plugin}
$ make all 
$ make install                    (to install as user to ~/bin/hivehub/...)
```

## Installation (draft)

The Hub is designed to run on Linux based computers. It might be able to work on other platforms but at this stage this is not tested nor a priority.

### System Requirements

The Hub can run on most small to large Intel and Arm based systems.

The minimal requirement for the Hub is 100MB of RAM and an Intel Celeron, or ARMv7 CPU. Additional resources might be required for some plugins. See plugin documentation.

### Install From Package Manager

Installation from package managers is currently not available.

### Manual Install As User

The Hub can be installed and run as a dedicated user or system user. This section describes to install the Hub in a dedicated user home directory.

0. Download or build the binaries. See the build section for more info.
1. Create a user, for example a 'hivehub' user. Login as that user.
2. Create the hub folder structure

```sh
mkdir -p ~/bin/hivehub/bin
mkdir -p ~/bin/hivehub/config
mkdir -p ~/bin/hivehub/logs 
mkdir -p ~/bin/hivehub/certs 
mkdir -p ~/bin/hivehub/certstore
```

3. Copy the application binaries into the ~/bin/hivehub/bin folder and default configuration in the ~/bin/hivehub/config folder

```sh
cp bin/* ~/bin/hivehub/bin
cp config/* ~/bin/hivehub/config
```

4. Generate the certificates using the certs CLI

```sh
cd ~/bin/hivehub
bin/certs certbundle   
```

5. Run

```sh
bin/launcher start
```

If desired, this can be started using systemd. Use the init/hivehub.service file.

### Install To System (tenative)

For systemd installation to run as user 'hivehub'. When changing the user and folders make sure to edit the init/hivehub.service file accordingly. From the dist folder run:

1. Create the folders and install the files

```sh
sudo mkdir /opt/hivehub/
sudo mkdir -P /etc/hivehub/certs/ 
sudo mkdir -P /var/lib/hivehub/certstore/ 
sudo mkdir /var/log/hivehub/   

# Install HiveOT configuration and systemd
# download and extract the binaries tarfile in a temp for and copy the files:
tar -xf hivehub.tgz
sudo cp config/* /etc/hivehub
sudo vi /etc/hivehub/hub.yaml    - and edit the config, log, plugin folders
sudo cp init/hivehub.service /etc/systemd/system
sudo cp bin/* /opt/hivehub
```

2. Setup the system user and permissions

```sh
sudo adduser --system --no-create-home --home /opt/hivehub --shell /usr/sbin/nologin --group hivehub
sudo chown -R hivehub:hivehub /etc/hivehub
sudo chown -R hivehub:hivehub /var/log/hivehub
sudo chown -R hivehub:hivehub /var/lib/hivehub

sudo systemctl daemon-reload
```

3. Start the hub

```sh
sudo service hivehub start
```

4Autostart the hub after startup

```sh
sudo systemctl enable hivehub
```

## Configuration

All Hub services will run out of the box with their default configuration. To change the default network and folder locations edit the 'config/hub.yaml' configuration file.

Hub services load their common configuration from the hub.yaml file in the config folder. This file MUST exist as it contains the message bus connection information for use by plugins. If no address is configured, the host outbound IP address is determined during startup. For hosts with multiple addresses, the address to use can be configured in hub.yaml

Plugins can have their own plugin specific configuration file in the config folder. Plugins must be able to run without a configuration file.

## Launching

The Hub can be launched manually by invoking the 'launcher' app in the Hub bin folder. eg:

```shell
~/bin/hivehub/bin/launcher
```

The services to start are defined in the config/launcher.yaml configuration file. When adding services, this file needs to be updated with the new service executable name.

A systemd launcher is provided that can be configured to launch on startup for systemd compatible Linux systems. See 'init/hivehub.service'

```shell
sudo cp init/hivehub.service /etc/systemd/system
sudo vi /etc/systmd/system/hivehub.service      (edit user and working directory)
sudo systemctl enable hivehub
sudo systemctl start hivehub
```

# Contributing

Contributions to HiveOT projects are always welcome. There are many areas where help is needed, especially with documentation and building plugins for IoT and other devices. See [CONTRIBUTING](docs/CONTRIBUTING.md) for guidelines.

# Credits

This project builds on the Web of Things (WoT) standardization by the W3C.org standards organization. For more information https://www.w3.org/WoT/

This project is inspired by the Mozilla Thing draft API [published here](https://iot.mozilla.org/wot/#web-thing-description). However, the Mozilla API is intended to be implemented by Things and is not intended for Things to register themselves. The HiveOT Hub will therefore deviate where necessary.

The [capnproto](https://capnproto.org/) project provides Capabilities based RPC infrastructure for the Hub. Capabilities based services are a great fit for a decentralized Hub as it is performant, low cpu and memory footprint and intrinsic secure.

Many thanks go to JetBrains for sponsoring the HiveOT open source project with development tools.  
