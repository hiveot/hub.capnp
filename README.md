# ARCHIVED NOTICE


# Hive-Of-Things Hub

The Hub for the *Hive of Things* provides a simple and secure base to view and operate IoT devices. The Hub securely mediates between consumers and IoT device 'Things' using a hub-and-spokes architecture. Consumers interact with Things via the Hub without connecting directly to the IoT devices or services. The Hub is based on the [W3C WoT TD 1.1 specification](https://www.w3.org/TR/wot-thing-description11/) and uses the [cap'n proto](https://capnproto.org/) for Capabilities based secure 
communication.


## Project Status

THIS PROJECT IS FUNCTIONAL BUT NOT LONGER FURTHER DEVELOPED. - April 2023

This experiment to build a microservices based IoT Hub has run its course. A couple of the takeaways:
* capn'proto is a solid choice for RPC. The capnp source files are very powerful, it is fast, and the golang compiler is robust. The capabilities based approach has the potential to make your application more secure. If you need a powerful RPC solution then I would recommend you consider this for your project. I'd like to thank the folks at go-capnproto2 for the excellent support.
* capnproto for javascript can use some love. Calling capnp API's from the browser requires to jump through some hoops, such as using wasm with the golang client. 
* RPC's are a lot of work. Especially maintaining an extensive API footprint with marshallers. It is overkill for this project.
* IoT at its roots is served better by a messaging platform. There is a lot of event generation that needs to be directed.  

Earlier iterations of hiveot were message based and found to be more suitable. Hence the decision is to move on to a message based approach. This RPC project will be renamed to hub.capnp and archived. The next message based iteration will be take the old name 'hub'.




Core micro-services of the Hub
```
- authn         user authentication management
- authz         user authorization of capabilities
- certs         certificate management for services
- directory     storage of the thing directory
- gateway       single entry point to obtain service capabilities for remote clients
                websocket support for Node clients and web browsers
- history       storage of thing event and action values -- needs rework
- hubcli        commandline interface to hub services for administrators
- launcher      starting and stopping of services
- mqtt          mqtt gateway to Hub pubsub, directory and history
- provisioning  automated provisioning of IoT devices using certificate authentication
- pubsub        publish and subscribe message bus for notifications between services
- resolver      local discovery and access to service capabilities
- state         easy to use persistance of state for services
```

For available protocol bindings see the [bindings repo](https://github.com/hiveot/bindings)

### Short Term Road Map
```
1. Dashboard viewer (hiveoview)
2. Bindings
   * isy99 Insteon binding (legacy protocol, for consideration)
   * weathermap
   * aurora watch
3. Client libraries using capnproto APIs:
   * javascript client for mqtt gateway
   * python client
4. Auto reconnect by clients
```


Additional Documention:
* [HiveOT Overview](https://hiveot.github.io/)
* [HiveOT design](docs/hiveot-design.md)
* [Thing TDs](docs/README-TD.md)


## Audience

This project is aimed at software developers and system implementors that are working on secure IoT solutions. HiveOT users subscribe to the security mandate that IoT devices should be isolated from the internet and end-users should not have direct access to IoT devices. Instead, all access operates via the Hub.

## Objectives

The primary objective of HiveOT is to provide a solution to secure the 'internet of things'.  

The state of security of IoT devices is appalling. Many of those devices become part of botnets once exposed to the internet. It is too easy to hack these devices and most of them do not support firmware updates to install security patches.

This security objective is supported by not allowing direct access to IoT devices and isolate them from the rest of the network. Instead, IoT devices discover and connect to a 'hub' to exchange information through publish and subscribe. Hub services offer 'capabilities' to clients via a 'gateway' proxy service. Capabilities based security ensures that capability can only be used for its intended purpose. 

> The HiveOT mandate is: 'Things Do Not Run (TCP) Servers'.

When IoT devices don't run TCP servers they cannot be connected to. This removes a broad attack surface. Instead, IoT devices connect to the hub using standard protocols for provisioning, publishing events, and subscribing to actions.

The secondary objective is to simplify development of IoT devices for the web of things. 

The HiveOT Hub supports this objective by handling authentication, authorization, logging, tracing, persistence, rate limiting, resiliency and user interface. The IoT device only has to send the TD document describing the things it has on board, submit events for changes, and accept actions by subscribing to the Hub.

The third objective is to follow the WoT and other open standard where possible.

Open standards improves interoperability with devices and 3rd party services. Protocol bindings provide this interop. 

Provide a decentralized solution. Multiple Hubs can build a bigger hive without requiring a cloud service and can operate successfully on a private network. 

HiveOT is based on the [W3C WoT TD 1.1 specification](https://www.w3.org/TR/wot-thing-description11/). See [docs/README-TD] for more information.


## Summary

Security is big concern with today's IoT devices. The Internet of Things contains billions of devices that when not properly secured can be hacked. Unfortunately the reality is that the security of many of these devices leaves a lot to be desired. Many devices are vulnerable to attacks and are never upgraded with security patches. This problem is only going to get worse as more IoT devices are coming to market. Imagine a botnet of a billion devices on the Internet ready for use by unscrupulous
actors.

This 'HiveOT Hub' provides capabilities to securely interact with IoT devices and consumers. This includes certificate management, authentication, authorization, provisioning, directory and history services.

HiveOT compatible IoT devices therefore do not need to implement these features. This improves security as IoT devices do not run Web servers and are not directly accessible. They can remain isolated from the wider network and only require an outgoing connection to the Hub. This in turn reduces required device resources such as memory and CPU (and cost). An additional benefit is that consumers receive a consistent user experience independent of the IoT device provider as all
interaction takes place via the Hub interface.

HiveOT follows the 'WoT' (Web of Things) open standard developed by the W3C organization, to define 'Things'. It aims to be compatible with this standard.

Integration with 3rd party IoT devices is supported through the use of protocol bindings. These protocol bindings translate between the 3rd device protocol and WoT defined messages.  

The communication infrastructure of the Hub is provided by 'Cap'n Proto', or capnp for short. Capnp provides a Capabilities based RPC for service invocation that is inherently secure. Only clients which have obtained a valid 'Capability' can invoke that capability, eg read a sensor or control a switch. The RPC will only pass requests that are valid, so the device does not have to concern itself with authentication and authorization. 

Since the Hub acts as the intermediary, it is responsible for features such as authentication, logging, resiliency, pub/sub and other protocol integration. The Hub can dynamically delegate some of these services to devices that are capable of doing so, potentially creating a decentralized solution that can scale as needed and recover from device failure. As a minimum the Hub manages service discovery acts as a proxy for capabilities. 

Last but not least, the 'hive' can be expanded by connecting hubs to each other through a 'bridge'. The bridge lets the Hub owner share select IoT information with other hubs.


## Build From Source

To build the hub and bindings from source, a Linux system with golang and make tools must be available on the target system. 3rd party plugins are out of scope for these instructions and can require nodejs, python and golang.

Prerequisites:

1. An x86 or arm based Linuxsystem. Ubuntu, Debian, Raspberrian 
2. Golang 1.19 or newer (with GOPATH set)
3. GCC Make any 2020+ version
4. Cap'n proto v0.8 or newer: https://capnproto.org/install.html  (only when rebuilding the capnproto API)
5. protobuf (snap)
  (only when rebuilding the gRPC API)

### Build Hub Services And CLI

NOTE: Capnp-v3 is in alpha and while pretty stable things might break at times.
Currently best to stick to 	capnproto.org/go/capnp/v3 v3.0.0-alpha.24

1. Download source code:
```sh
git clone git@github.com:hiveot/hub
cd hub
``` 
2. Install go-capnp tools 
```sh
sudo apt get install capnproto (for Ubuntu)
make setup 
```
3. Build the hub
```sh
make hub
```

After the build is successful, the distribution files can be found in the 'dist' folder.



### Build Bindings

See the README of the bindings repository. 
In short:
```bash
git clone https://github.com/hiveot/bindings
cd {plugin}
make all 
```

## Install To User 

To install and run the Hub as the current user to ~/bin/hiveot.

```sh
make install
```

This copies the distribution files to ~/bin/hiveot. The method can also be used to upgrade an existing installation. Executables are always replaced but only new configuration files are installed. Existing configuration remains untouched to prevent wrecking your working setup.


### Uninstall:
To uninstall simply remove the ~/bin/hiveot folder.


## Install To System (tenative)

While it is a bit early to install hiveot as a system application, this is how it could work:

For systemd installation to run as user 'hiveot'. When changing the user and folders make sure to edit the init/hiveot.service file accordingly. From the dist folder run:

1. Create the folders and install the files

```sh
sudo mkdir -P /opt/hiveot/services/bindings
sudo mkdir -P /etc/hiveot/conf.d/ 
sudo mkdir -P /etc/hiveot/certs/ 
sudo mkdir /var/log/hiveot/   
sudo mkdir /var/lib/hiveot   
sudo mkdir /run/hiveot/

# Install HiveOT 
# download and extract the binaries tarfile in a temp for and copy the files:
tar -xf hiveot.tgz
sudo cp config/* /etc/hiveot/conf.d
sudo vi /etc/hiveot/hub.yaml    - and edit the config, log, plugin folders
sudo cp -a bin/* /opt/hiveot
```

Add /opt/hiveot/bin to the path

2. Setup the system user and permissions

```sh
sudo adduser --system --no-create-home --home /opt/hiveot --shell /usr/sbin/nologin --group hiveot
sudo chown -R hiveot:hiveot /etc/hiveot
sudo chown -R hiveot:hiveot /var/log/hiveot
sudo chown -R hiveot:hiveot /var/lib/hiveot
```

## Docker Installation

This is planned for the future with the beta release.

## Configuration

All Hub services will run out of the box with their default configuration. Service can use an optional yaml based configuration file found in the config folder.

### Generate a CA certificate
Before starting the hub, a CA certificate must be created. By default, the hub uses a self-signed CA certificate. It is possible to use a CA certificate from a 3rd party source, but this isn't needed as the certificates are used for client authentication, not for domain verification.

Generate the CA certificate using the CLI:

```sh
cd ~/bin/hiveot        # when installed locally
bin/hubcli ca create   # or simply "hubcli ca create" when the path is set 
```

### Service Autostart Configuration
To configure autostart of services edit the provided launcher.yaml and add the services to the autostart section.
> vi config/launcher.yaml

### Systemd Configuration

Automatic startup after boot is supported through a systemd service. This can be used when installed system wide or as a user. 


```shell
vi init/hivehub.service    #  (edit user, group, paths)
sudo cp init/hivehub.service /etc/systemd/system
sudo systemctl daemon-reload
sudo systemctl enable hivehub
sudo systemctl start hivehub
```

Once running, the running services can be viewed using the hub cli:
> hubcli launcher list

To stop or start a service:
> hubcli launcher stop {serviceName}

> hubcli launcher start {serviceName}

# Contributing

Contributions to HiveOT projects are always welcome. There are many areas where help is needed, especially with documentation, testing and building bindings for IoT and other devices. See [CONTRIBUTING](CONTRIBUTING.md) for guidelines.

# Credits

This project builds on the Web of Things (WoT) standardization by the W3C.org standards organization. For more information https://www.w3.org/WoT/

This project is inspired by the Mozilla Thing draft API [published here](https://iot.mozilla.org/wot/#web-thing-description). However, the Mozilla API is intended to be implemented by Things and is not intended for Things to register themselves. The HiveOT Hub will therefore deviate where necessary.

The [capnproto](https://capnproto.org/) project provides Capabilities based RPC infrastructure for Hub services. Capabilities based services are a great fit for a decentralized Hub as it is performant, low cpu and memory footprint and intrinsic secure.

Many thanks go to JetBrains for sponsoring the HiveOT open source project with development tools.  
