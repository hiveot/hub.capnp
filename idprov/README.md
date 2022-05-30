# IoT Device Provisioning

## Objective

Provides a simple means to securely provision IoT devices using out-of-band device verification and
bulk provisioning. This is a golang implementation of
the '[idprov standard](https://github.com/wostzone/idprov-standard)'

Provisioned devices receive a signed client certificate that can be used to authenticate themselves
with the WoST Hub and its services.

## Summary

This project implements the 'idprov' IoT device provisioning protocol. It provides a client (in
hubclient-go/idprovclient) and server library, and commandline utilities for standalone operation.
It is intended for automated provisioning of IoT devices to enable secure authenticated connections
from an IoT device to IoT services.

The typical use-case is that upon installing one or more IoT devices, the administrator collects the
device ID and corresponding out-of-band secret and provides these to the provisioning server using
the commandline utility or (future) web interface. When the devices are powered on the following
process takes place:

1. The IoT device discovers the provisioning server on the local network using DNS-SD )
2. The IoT device requests a certificate from the provisioning server providing device identity and
   a hash of the out of band secret.
3. The provisioning server verifies the device identity by matching the device-ID and secret with
   the administrator provided information. If there is a match then the device is issued a signed
   identity certificate.
4. The certificate is then used by the device to authenticate itself with IoT service providers,
   publish its information, and receive actions and configuration updates.

IoT devices can use the provided client library (hubclient-go) to implements this process in a few
lines of code. If no special OOB secret is available or needed, the device MAC address can be used
as secret. The client ID can be the device's hostname, serial number or dedicated ID.

The protocol uses the organizational unit (ou) field of the certificate to assign devices to the
organization of IoT devices, with corresponding permissions.

This project provides:

The idprov project provides:

1. The ['idprov-standard'](https://github.com/wostzone/idprov-standard)) provisioning protocol
   definition.
2. A ['client library'](https://github.com/wostzone/hub/tree/main/idprov/pkg/idprovclient) for IoT
   devices to obtain a certificate.
3. The ['provisioning server'](https://github.com/wostzone/hub/tree/main/idprov/pkg/idprovserver)
   reference implementation for issuing signed certificates to IoT devices.
4. An [out-of-band commandline utility](https://github.com/wostzone/hub/tree/main/idprov/cmd/oob) (
   cmd/oob) utility for posting out of band secrets needed for provisioning.

## Features

This server supports the following features:

1. Publish DNS-SD discovery record for the idprov server
2. Set out of band secret for devices
3. Issuing device certificates signed by the CA
4. Provide a directory with the endpoints for provisioning
5. Provide a list of services with addresses and ports that accept the client certificate

## Usage

On startup, the IDProv server publishes a DNS-SD record on the local network. IoT devices can
discovery it using the idprov client 'discover' function. Alternatively, IoT devices are provided
directly with the server address and port.

Once the idprov server is discovered, devices obtain the services directory and submit a
provisioning request including their ID, out-of-band secret and public key. Once approved the server
returns a certificate that is stored by the device and used for TLS connections with other Hub
services. Periodically the Device renews the certificate by submitting a provisioning request
halfway the existing certificate validity period.

In order to receive a certificate, the device ID and secret must be submitted out of band to the
server before the devices requests provisioning. This can be done using the oob utility or via the
Hub's admin UI if available. The admin UI shows a list of requests which the administrator can
approve. Devices must retry repeatedly if their request returns the status 'waiting'.

If no special OOB secret is available, devices can use their MAC address as ID and serial number as
its secret. This is up to the device itself. The easiest method for provisioning is the use of QR
code or NFC tag on the device that can be scanned. A provisioning app can automatically pass this on
as out-of-band verification to the server. For bulk provisioning a list of IDs and secrets can be
provided using the oob utility.

The certificate provided in provisioning to the thing device must be used in order to connect
securely to any of the Hub services that are listed in the 'get directory' request, such as the MQTT
message bus, or other Hub services. All connections must use mutual authentication over TLS to
abtain sufficient permissions.

In this example the 'clientCertFolder' is the folder where the client stores its public/private
key-pair and the issued certificates. The public/private key-pair is created on first start if no
existing key is found.

(This code is an example only and won't work unless a running server is available)

```golang
import "github.com/wostzone/hub/idprov/pkg/idprovclient"

func provisionMe() error {
// Create instance of the client
// Use the MAC as unique deviceID and serialnr as oob secret. 
myDeviceID := "AA:BB:CC:DD:EE:FF"
secret := "12345678"
// use a secure location to store certificate info
certPath := "./clientcerts/cert.pem"
keyPath := "./clientcerts/key.pem"
caCertPath := "./clientcerts/caCert.pem"
// without server address, discovery will be used to locate it on the local network 
idpClient := idprovclient.NewIDProvClient( myDeviceID, "", certPath, keyPath, caCertPath)

// Connect to the server and obtain the provisioning directory and CA certificate
// This also creates a ECDSA public/private key-pair if they don't yet exist
err := idpClient.Start()
if err != nil {
return err
}

// Request a certificate using the out-of-band secret
// Certificate is stored in the provided certPath location
// If a certificate already exists it is refreshed
caCert, myCert, err := idprov.Provision(deviceID, secret)
if err != nil {
return err
}
}
```

Next, the client uses the certificate connecting to the hub message bus and publish messages:

```golang
    import "github.com/wostzone/wost-go/pkg/certs"

// The provisioned services list can be queried from the provisioning server on startup. 
mqttAddress := idprov.GetService("mqtt")

myCert := certs.LoadTLSCertFromPEM(certPath, keyPath)
caCert := certs.LoadX509CertFromPEM(caCertPath)
mqttHubClient := NewMqttHubClient(deviceID, caCert)
mqttHubClient.ConnectWithClientCert(mqttAddress, myCert)
...
mqttHubClient.PublishTD(myThingDescription)
...
mqttHubClient.Stop()
```

## Installation

This [plugin] is included with the Hub as a core service and is bundled with the Hub installation.

## Configuration

This service is enabled by default in the launcher.yaml configuration file that lists the plugins to
run on startup. To disable this plugin simply comment-out the protocol.

The service can be configured using the idprov.yaml configuration file. An example is available in
the Hub's config folder. If no configuration file is available the server will be started with
default values.

The plugin configuration allows for:

* configure the listening address and port of the idprov server
* enable/disable discovery publications on the local network
* set the logging level for the plugin

See the config/idprov.yaml file for more detail.
