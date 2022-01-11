# IoT Device Provisioning

This project provides a golang implementation for the '[idprov standard](https://github.com/wostzone/idprov-standard)'
IoT device provisioning server. The protocol describes how to issue signed certificates to IoT devices with support for
out-of-band verification and bulk provisioning.

The signed certificates are for use by IoT devices to make secure authenticated connections to service providers.

## Project Status

Status: Alpha

The status of this library is alpha. It is functional but breaking changes can be expected.

## Audience

This project is aimed at IoT developers that need a method of provisioning IoT devices with support for out-of-band
verification. 'WoST' developers choose not to run servers on Things and instead use a hub and spokes model.

## Summary

This project implements the 'idprov' IoT device provisioning protocol. It provides a client (in
hubclient-go/idprovclient) and server library, and commandline utilities for standalone operation. It is intended for
automated provisioning of IoT devices to enable secure authenticated connections from an IoT device to IoT services.

The typical use-case is that upon installing one or more IoT devices, the administrator collects the device ID and
corresponding out-of-band secret and provides these to the provisioning server using the commandline utility or (future)
web interface. When the devices are powered on the following process takes place:

* The IoT device discovers the provisioning server on the local network using DNS-SD )
* The IoT device requests a certificate from the provisioning server providing device identity and a hash of the out of
  band secret.
* The provisioning server verifies the device identity by matching the device-ID and secret with the administrator
  provided information. If there is a match then the device is issued a signed identity certificate.
* The certificate is then used by the device to authenticate itself with IoT service providers, publish its information,
  and receive actions and configuration updates.

IoT devices can use the provided client library (hubclient-go) to implements this process in a few lines of code. If no
special OOB secret is available or needed, the device MAC address can be used as secret. The client ID can be the
device's hostname, serial number or dedicated ID.

The protocol uses the organizational unit (ou) field of the certificate to assign devices to the organization of IoT
devices, with corresponding permissions.

This project provides:

1. The ['idprov-standard'](https://github.com/wostzone/idprov-standard)) provisioning protocol definition
2. A ['client library'](https://github.com/wostzone/hub/idprov/pkg/idprovclient) 
3. The ['provisioning server'](https://github.com/wostzone/hub/idprov/pkg/idprovserver)
4. An [out-of-band commandline utility](https://github.com/wostzone/hub/idprov/cmd/oob) (cmd/oob)

## Features

This server supports the following features:

1. Publish DNS-SD discovery record for the idprov server
2. Set out of band secret for devices
3. Issuing device certificates signed by the CA
4. Provide a directory with the endpoints for provisioning
5. Provide a list of services with addresses and ports that accept the client certificate

## Usage

The hub/lib/client library makes server discovery and IoT device provisioning a simple endeavor. (other languages are planned)

In this example the 'clientCertFolder' is the folder where the client stores its public/private key-pair and the issued certificates. The public/private key-pair is created on first start if no existing key is found.

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
    import "github.com/wostzone/hub/lib/client/pkg/certs"

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
