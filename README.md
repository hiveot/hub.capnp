# Hive-Of-Things Hub

The Hub for the *Hive of Things* is an intermediary between IoT devices 'Things', services, and consumers using a hub-and-spokes architecture. Consumers interact with Things via the Hub without connecting directly to the Thing device. The Hub uses the [cap'n proto](https://capnproto.org/) for Capabilities based secure communication.

## Project Status

Status: The status of the Hub is In Development. It is undergoing a rewrite to use **capnp** for infrastructure.

## Audience

This project is aimed at software developers and system implementors that are working on secure IoT devices. Users choose to not run servers on Things and instead use a hub and spokes model which greatly reduces the security risk post by traditional IoT devices.

## Objectives

1. The primary objective of HiveOT is to provide a solution to secure the 'internet of things'.  

The state of security of IoT devices is appalling. Many of those devices become part of botnets once exposed to the internet. It is too easy to hack these devices and most of them do not support firmware updates.

The security objective is supported by not allowing direct access to IoT devices and isolate them from the rest of the network. Instead, IoT devices discover and connect to a 'hub' to exchange information through publish and subscribe. Hub services offer 'capabilities' to clients via a 'gateway' proxy service. Capabilities based security ensures that capability can only be used for its intended purpose. Unlike authentication tokens which when compromised offer access to all services of the 
user.

> The HiveOT mandate is: 'Things Do Not Run (TCP) Servers'.

2. The secondary objective is to simplify development of IoT devices for the web of things. 

The HiveOT supports this objective by handling authentication, authorization, logging, tracing, persistence, rate limiting and resiliency. The IoT device only has to send the TD of things it has on board, submit events for changes, and accept actions by subscribing to the Hub.

3. The third objective is to follow the WoT and other open standard where possible.

HiveOT is based on the [W3C WoT TD 1.1 specification](https://www.w3.org/TR/wot-thing-description11/). See [docs/README-TD] for more information.


## Summary

This document describes a high level overview of the Hub.

Security is big concern with today's IoT devices. The Internet of Things contains billions of devices that when not properly secured can be hacked. Unfortunately the reality is that the security of many of these devices leaves a lot to be desired. Many devices are vulnerable to attacks and are never upgraded with security patches. This problem is only going to get worse as more IoT devices are coming to market. Imagine a botnet of a billion devices on the Internet ready for use by unscrupulous
actors.

This 'HiveOT Hub' provides core services to securely interact with IoT devices and consumers. This includes certificate management, authentication, authorization, provisioning, message bus service and directory service.

HiveOT compatible IoT devices (Things) therefore do not need to implement these features. This improves security as IoT devices do not run Web servers and are not directly accessible. They can remain isolated from the wider network and only require an outgoing connection to the Hub. This in turn reduces required device resources such as memory and CPU (and cost). An additional benefit is that consumers receive a consistent user experience independent of the IoT device provider as all
interaction takes place via the Hub interface.

HiveOT follows the 'WoT' (Web of Things) open standard developed by the W3C organization, to define 'Things'. It aims to be compatible with this standard.

The communication infrastructure for the services is provided by 'Cap'n Proto', or capnp for short. Capnp provides a Capabilities based RPC for service invocation that is inherently secure. Only clients which have obtained a valid 'Capability' can invoke that capability, eg read a sensor or control a switch. The RPC will only pass requests that are valid, so the device does not have to concern itself with authentication and authorization. 

Since the Hub acts as the intermediary, it is responsible for features such as authentication, logging, resiliency, pub/sub and other protocol integration. The Hub can dynamically delegate some of these services to devices that are capable of doing so, potentially creating a decentralized solution that can scale as needed and recover from device failure. As a minimum the Hub manages service discovery acts as a proxy for capabilities. 


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
