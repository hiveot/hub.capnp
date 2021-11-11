# WoST Hub Server Library

This repository provides a library with definitions and methods to provide services as part of the WoST Hub. For developing clients of the Hub see 'hubclient-go'.

## Project Status

Status: The status of this library is Alpha. It is functional,and has a test coverage of over 90%. However, breaking changes must be expected.

Under consideration:
* Signing of messages is under consideration. Most likely using JWS.
* Encryption of messages. Presumably using JWE. It can be useful for sending messages to the device that should not be accessible to others on the message bus.

## Audience

This repository is intended for developers of services for the WoST Hub. WoST Hub services follow the paradigm that Things do not run servers. Hub Services are servers that are secure and can be upgraded over the air using the Hub upgrader.


## Summary

This library provides functions for creating WoST Hub services. Developers can use it to create a secure TLS server that:
- manage certificates for clients
- works with the Hub authentication mechanism using client certificate authentication and username/password authentication. 
- provides DNS-SD discovery of the service
- watch configuration files for changes

A Python and Javascript version is planned for the future.
See also the [WoST Hub client library](github.com/wostzone/hubclient-go) for additional client oriented features. 

## Dependencies

This requires the use of a WoST compatible Hub or Gateway.  

Supported hubs and gateways:
- [WoST Hub](https://github.com/wostzone/hub)


## Usage

This module is intended to be used as a library by Hub services developers. 

### Service Configuration

Services can be configured using yaml files. The client config library provides a standard way to
load configuration and receive notifications if configuration files are changed. 
See github.com/wostzone/hubclient-go/pkg/config for details.

The standard folder structure for a service is as follows:
```
/home/wost/bin/hub/        the hub app folder
                |- bin     plugin binaries
                |- config  configuration folder
                |- certs   certificates
                |- logs    logging files
```

The global hub.yaml contains Hub configuration settings with folders to use. If the above directory
structure is not usable for whatever reason, the hub.yaml file can change the default folder structure.


```golang
	import "github.com/wostzone/hubclient-go/pkg/config"
  ...
  // Load the service and hub configuration
  var myconfig MyServiceConfig{}
	hubConfig, err := config.LoadConfig(homeFolder, pluginID, &myconfig)
  ...
```

### certsetup

The certsetup package provides functions for creating, saving and loading self signed certificates include a self signed Certificate Authority (CA). These are used for verifying authenticity of server and clients of the message bus.


### tlsserver

Server of HTTP/TLS connections that supports certificate and username/password authentication, and authorization.

Used by the IDProv protocol server and the Thingdir directory server.


# Contributing

Contributions to WoST projects are always welcome. There are many areas where help is needed, especially with documentation and building plugins for IoT and other devices. See [CONTRIBUTING](https://github.com/wostzone/hub/docs/CONTRIBUTING.md) for guidelines.


# Credits

This project builds on the Web of Things (WoT) standardization by the W3C.org standards organization. For more information https://www.w3.org/WoT/
