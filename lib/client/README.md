# WoST Hub Client Library 

This repository provides a library with definitions and methods to use WoST Hub services. It is intended for developing IoT "Thing" devices and for developing consumers of Thing information.

## Project Status

Status: The status of this library is Alpha. It is functional,and has a test coverage of over 90%. However, breaking changes must be expected.

## Audience

This repository is intended for developers of IoT devices and for developers of applications using WoST IoT 'Thing' information. WoST users follow the paradigm that Things do not run servers and communication with IoT devices flows via the WoST secure Hub.

## Summary

This Go library provides common building blocks for creating WoST Hub clients such as IoT devices, protocol bindings and consumers. 

## config



## discovery

The discovery package lets clients discover of services by their service name. This is used for example in the idprov provisioning client to discover the provisioning server. 

For example, to discover the URL of the idprov service:

```golang
   serviceName := "idprov"
   address, port, paraMap, records, err := discovery.DiscoverServices(serviceName, 0)
```

## idprovclient

The idprov client is used by Thing Devices to discovery the idprov provisioning server. This server provides the client with a certificate after confirmation of an out-of-band secret.

This services can also used by other users to discovery the mqtt message bus through the IDProv directory.

## mqttclient

Client to connect to the Hub MQTT broker. The MQTT client is build around the paho mqtt client and adds reconnects, and CA certificate verification with client certificate or username/password authentication.

The MqttHubClient includes publishing and subscribing to WoST messages such as Action, Config (properties), Events, Property value updates and the full TD document. WoST Thing devices use these to publish their things and listen for action requests.

For example, to connect to the message bus using a client certificate:
```golang
   serviceName := "idprov"
   address, port, paraMap, records, err := discovery.DiscoverServices(serviceName, 0)
```
## signing

>### This section is subject to change
The signing package provides functions to JWS sign and JWE encrypt messages. This is used to verify the authenticity of the sender of the message.

Signing support is built into the HTTP and MQTT protocol binding client and server. 
Well, soon anyways... 

Signing and sender verification is key to guarantee that the information has not been tampered with and originated from the sender. The WoT spec does not (?) have a place for this, so it might become a WoST extension.


## tlsclient

TLSClient is a client for connecting to TLS servers such as the Hub's core ThingDirectory service. This client supports both certificate and username/password authentication using JWT with refresh tokens.

For example, a client device can connect to a Hub service as follows
```golang
  caCert := LoadCertFromPem(pathToCACert)
  clientCert := LoadCertFromPem(pathToClientCert)
	client, err := tlsclient.NewTLSClient("host:port", caCert)
	err = client.ConnectWithClientCert(clientCert)
  ... do stuff ...
  client.Close()
```

## td

Helper functions to build a Thing Description, Action, Event and Configuration documents for publishing on the message bus.

Note: The generated TD is basic and might not conform to the WoT standard that is itself in flux (June 2021)


For example, to build a new TD of a temperature sensor Thing:
```golang
	import "github.com/wostzone/hubclient-go/pkg/td"
	import  "github.com/wostzone/hubclient-go/pkg/vocab"

  ...

  thing := td.CreateTD("thingID1", vocab.DeviceTypeSensor)
  versions := map[string]string{"Software": "v10.1", "Hardware": "v2.0"}
  td.SetThingVersion(thing, versions)

 	prop := td.CreateProperty("otemp", "Outdoor temperature", vocab.PropertyTypeSensor)
	td.SetPropertyUnit(prop, "C")
	td.SetPropertyDataTypeInteger(prop, -100, 100)
	td.AddTDProperty(thing, "temperature", prop)
```

Under consideration:
* Signing of messages. Most likely using JWS.
* Encryption of messages. Presumably using JWE. It can be useful for sending messages to the device that should not be accessible to others on the message bus.


### testenv

The 'testenv' package includes a mosquitto launcher for testing clients including plugins.

## vocab

Ontology with vocabulary used to describe Things. This is based on terminology from the WoT working group and other source. When no authorative source is known, the terminology is defined as part of the WoST vocabulary. 

This includes devicetype names, Thing property types, property names, unit names and TD defined terms for describing a Thing Description document.

## testenv

testenv simulates a server for testing of clients. This includes generating of certificates and setup and run a mosquitto mqtt test server.

For example, to test a client with a mosquitto server using the given configuration and certificate folder for use by mosquitto:
```golang
	certs = testenv.CreateCertBundle()
	mosquittoCmd, err := testenv.StartMosquitto(configFolder, certFolder, &certs)
  ...run the tests...
	testenv.StopMosquitto(mosquittoCmd)
```
See: pkg/mqttclient/MqttClient_test.go for examples

# Contributing

Contributions to WoST projects are always welcome. There are many areas where help is needed, especially with documentation and building plugins for IoT and other devices. See [CONTRIBUTING](https://github.com/wostzone/hub/docs/CONTRIBUTING.md) for guidelines.


# Credits

This project builds on the Web of Things (WoT) standardization by the W3C.org standards organization. For more information https://www.w3.org/WoT/
