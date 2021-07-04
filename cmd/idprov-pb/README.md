# IoT Device Provisioning - Protocol Binding 

This project implements the protocol binding for the ['idprov protocol'](https://github.com/wostzone/idprov-standard) using the [idprov library](https://github.com/wostzone/idprov-go).

This protocol binding lets 'Things' discover the idprov server on the local network and obtain a certificate to be able to connect to Hub services like the messaging bus. 

## Project Status

Status: Alpha

The status of this library is alpha. It is functional but breaking changes should be expected.

## Audience

This protocol binding is aimed at IoT developers that need a method of provisioning IoT devices with support for out-of-band verification. 'WoST' developers choose not to run servers on Things and instead use a hub and spokes model. 

## Summary

This protocol binding is part of the Hub core. It provides the means for Things to discover the Hub on the local network and authenticate with it using out-of-band verification. 

The protocol binding starts the idprov server which publishes a DNS-SD record on the local network. IoT devices can discovery it using the idprov client 'discover' function. Alternatively, IoT devices are provided with the server address and port.

Once the idprov server is discovered, devices obtain the services directory and submit a provisioning request including their ID, out-of-band secret and public key. Once approved the server returns a certificate that is stored by the device and used in for TLS connections with other Hub services. Periodically the Device renews the certificate by submitting a provisioning request halfway the existing certificate validity period.

In order to be approved the device ID and secret must be submitted to the server before the provisioning request. This can be done using the oob utility or via the Hub's admin UI if available. Devices must retry repeatedly if their request returns the status 'waiting'.

If no special OOB secret is available devices can use their MAC address as ID and  serial number as its secret. This is up to the device itself. The easiest method for provisioning is the use of QR code or NFC tag on the device that can be scanned with a phone. A provisioning app can automatically pass this on as out-of-band verification to the server.

The provided certificate must be used to connect securely to any of the Hub services that are listed in the 'get directory' request, such as the MQTT message bus, or other Hub services. All connections must use mutual authentication over TLS to abtain sufficient permissions. 

The auth plugin is used to authorize access to device Things by consumers.

The idprov project provides:
1. The ['idprov-standard'](https://github.com/wostzone/idprov-standard)) provisioning protocol definition
2. A ['client library'](https://github.com/wostzone/idprov-go/pkg/idprov) for IoT devices to obtain a certificate.
3. The ['provisioning server'](https://github.com/wostzone/idprov-go/pkg/idprovserver)'  reference implementation for issuing signed certificates to IoT devices.
4. An [out-of-band commandline utility](https://github.com/wostzone/idprov-go/pkg/idprov-oob) utility for posting out of band secrets needed for provisioning.


## Installation

This protocol binding is included in the Hub as a core protocol. It is installed automatically with the Hub installation.

## Configuration

This protocol is enabled by default in the hub.yaml configuration file that lists the plugins to run on startup. To disable this plugin simply comment-out the protocol.

The protocol binding can be configured using the idprov-pb.yaml configuration file. An example is available in the Hub's config folder. If no configuration file is available the server will be started with default values.

The plugin configuration allows for:
* configure the listening address and port of the idprov server
* enable/disable discovery publications on the local network
* set the logging level for the plugin

See the config/idprov-pb.yaml file for more detail.

## Dependencies

This library uses the wostlib-go and idprov-go projects for managing certificates, start TLS client and server connections, and run the idprov server.

Clients will need to use the idprov-go/pkg/idprov client library to discover the server and obtain authentication certificates. 
