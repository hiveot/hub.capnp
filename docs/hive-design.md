# Hive Of Things Design

![Design Overview](./hub-overview.png)

## Overview

The HiveOT design follows a 'hub-and-spokes' architecture, where IoT devices and other clients connect to the nearest Hub to access the Hive's capabilities.

The Hive's Hub capabilities provides a Thing directory, history, current values and much more. These capabilities are provided by services that run on the Hive's computing devices and are accessed through the computing device's gateway. Every computing device in the Hive has such a gateway. The gateways communicate with each other to share their capabilities.

At no point do IoT devices and consumers connect to each other directly unless this is explicitly by design, like for example a media server.

## Design patterns
The following design patterns are leveraged:

* [Hub and Spokes with network peering](https://cloud.google.com/architecture/deploy-hub-spoke-vpc-network-topology) centralizes access to IoT devices via a central Hub. In this case the Hub itself can be distributed using network peering. This provides isolation between IoT devices, services, and consumers. As IoT devices are notoriously insecure this establishes a secure wall between them and the outside world.
* [Microservices architecture](https://docs.microsoft.com/en-us/dotnet/architecture/microservices/architect-microservice-container-applications/microservices-architecture) support the single responsibility principle, simplifies testing and reduces the risk of bugs by focusing on a single task.
* [Capabilities based security](https://en.wikipedia.org/wiki/Capability-based_security) "A capability is a communicable, unforgeable token of authority." 
* [API Gateway pattern](https://docs.microsoft.com/en-us/dotnet/architecture/microservices/architect-microservice-container-applications/direct-client-to-microservice-communication-versus-the-api-gateway-pattern) provides a single endpoint for clients to communicate with the services instead of a separate connection to each service. This reduces coupling with the services and enables the use of middleware for common tasks such as authentication.
* [BFF design pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/backends-for-frontends) (Front For Backend Service). Dedicated services tailored to front-end usage separate the internal services API from the external API. Internal design changes do not affect the APIs and vice versa. This approach further reduces the number of request round trips as the API is tailored to the client's use-case.

Even though a distributed architecture might seem overkill for a simple solution, it opens the door to a resilient decentralized approach where multiple computing units can provide the services without the need of a central coordinating single point of failure. Any computing device can participate to run services, increasing the capacity of the system with little or no dedicated hardware. 

## Infrastructure

The Hub infrastructure is build on [Cap'n Proto](https://capnproto.org/) and leverages [capabily based security](https://en.wikipedia.org/wiki/Capability-based_security). Cap'n proto can use various types of transport, like a simple pipe, Unix domain sockets, and TCP sockets. Communication is encrypted using TLS.

Internally, Hub services communicate using capnp over Unix Domain sockets. Externally, the Hub Gateways provide capnp APIs using TCP sockets.  

A gateway services provides capabilities to the services that are available on the Hub itself or via the gateway of a participating device. While the extra hop to a different gateway might add extra latency, the (planned) use of [level 3 protocol](https://capnproto.org/rpc.html) will allow the capability to be transferred using a direct connect to the offering service. Level 3 is in development for the go-capnp implementation in 2022.

If identical capabilities are available from multiple gateways then the connected gateway automatically selects the most efficient option. This allows for automatic failover in case a gateway is no longer available.

Access to capabilities via non-capability protocols, such as HTTP, MQTT, gRPC, is provided through so-called protocol gateway services. These services provide a bridge between the external facing protocol and the internal capnp protocol. In order to access these services authorization is required, based on the gateway protocol. For example BASIC, DIGEST, or OAUTH2 authentication.

While a Gateway provides the ability to access capabilities, the available capabilities are constrained by authorization of the client. Client authorization can be obtained by using a client certificate or login credentials for authentication.

To determine which capability an authenticated client has access to, the gateway uses the authorization service. The rules are based on the type of client, eg IoT device, service, or user, and the role in groups the client shares with Things.




## IoT Gateway

The HiveOT includes various gateways to facilitate communication between the Hub and the world around it. The IoT Gateway is intended for interaction with IoT devices and protocol adapters.

The IoT Gateway service is a capnp based service that provides capabilities for IoT devices to:
1. Discover the IoT Gateway endpoint(s) to use.
2. Provision an IoT device with a valid certificate.
3. Publish the TD documents of Things it is responsible for.
4. Publish Thing events.
5. Subscribe to Thing action requests.

IoT devices that support this API can use it directly. The capabilities API provide inherit security through built-in constraints. Only capabilities that the client is allowed to use are provided from the IoT Gateway. 

See the IoT gateway service README for more details on using its API.

### IoT Gateway Protocol Bindings

IoT Protocol bindings are services that translate between a 3rd party IoT protocol and the IoT Gateway API. The following bindings are on the short to intermediate roadmap:

1. HTTPS 'idprov' binding that implements the idprov discovery and provisioning protocol.
2. HTTPS binding for publishing TD documents and events.
3. HTTPS Websocket binding for subscribing to thing actions.
4. ZWave binding that connects to a zwave USB controller and is a publisher for its ZWave devices.
5. OwServer one-wire binding that connects to a OWServer gateway is a publisher for its 1-wire devices.
6. Openweathermap binding that connects to the OWM server and is a publisher for weather information.
7. IPCam binding that is a publisher for IP Cameras snapshots.
8. ISY99x binding that is a publisher for ISY99x connected Insteon devices.
9. SNMP binding that is a publisher for SNMP discovered network devices.
10. Location tracking bindings:
    a. Bluetooth location binding that discovers nearby bluetooth devices.
    b. Wifi location binding that discovers nearby wifi mobile devices.
    c. Android location tracking service that tracks Android user location.
    d. iPhone location tracking service that tracks iPhone user location.
    e. Bindings for integration with location tracking services.
    f. Image recognition to identify people and animals.
11. ZigBee binding that is a publisher for ZigBee network devices.
12. Hue binding that connects to Philips Hue lights.
13. LoRa binding connects to a LoRa gateway and is a publisher for LoRa devices.
14. CoAP binding that connects to a CoAP gateway is a publisher for CoAP devices.
15. Health and Activity tracking bindings that determine what is going on:
    a. Fitbit integration binding.
    b. Emergency button binding for detecting an alarm. - senior living safety, lone worker safety, ...

There are hundreds more potential bindings to develop. In order to simplify development a library supporting the WoT scripting API is (will be) available in golang, javascript, python and other programming languages.

## User Gateway

The user gateway is intended for interacting with end-users of information collected by the HiveOT Hub.

The user gateway service is a capnp based service that provides capabilities for end users to:
1. Discover the User Gateway endpoint(s) to use.
2. Authenticate a user.
3. Retrieve available Things from the directory. 
4. Read historical Thing values from the history store.
5. Subscribe to Thing value updates.
6. Publish Thing actions.

See the User gateway service README for more details on using its API.

### User Gateway Protocol Bindings

User gateway protocol bindings are services that translate between the Hub's capnp protocol and end-user protocols. The following bindings are on the short to intermediate roadmap:

1. DNS-SD to discover gateway addresses
2. HTTPS Authentication using BASIC, DIGEST, OAUTH2
3. HTTPS REST API to query Thing directory and values
4. MQTT binding to pub/sub over MQTT


### Admin Gateway

The admin gateway is intended for allowing administrators to manage the HiveOT Hub.

The admin gateway is a capnp based service that provides capabilities to:
1. List provisioning requests.
2. Approve or revoke a provisioning request.
3. Manage users (CRUD)
4. Manage groups of Things and Users
5. Manage certificates

See the Admin gateway service README for more details on using its API.

### Pub/Sub Communication

Consumers subscribe to events and can publish actions, while IoT devices or services publish events and subscribe to actions. For compatible clients this takes place using capnproto streams.

Consumers that do not support the capnp protocol have the alternative of using one of the available protocol adapters. Planned are MQTT and Websockets. 


### Intermittent Connectivity

A limitation of network devices is that they only communicate when awake and connected. Battery operated devices might spend most of their time asleep while remote devices might suffer from intermittent connectivity.

The Hive detects device connectivity and updates the status accordingly. On reconnect the device will receive queued any actions. IoT devices will have to queue their outgoing messages when disconnected, and send them when connection is reestablished.

TBD: who tracks, directory service or history service? Where is this persisted?

## Core Services

Core services provide necessary capabilities to be able to function as a Hub. Access to these capabilities is provided by the 'gateway' service described above. 

### Provisioning

Provisioning is the process of pairing an IoT device to the Hub. 

IoT devices that support the [idprov protocol](https://github.com/hiveot/idprov-standard) can automatically discover the Hub on the local network using the DNS-SD protocol and initiate the provisioning process. When accepted, a CA signed device (client) certificate is issued.

The device certificate supports machine to machine authentication between IoT device and Hub. See [idprov service](https://github.com/hiveot/hub/tree/main/idprov) for more information.

### Authentication (authn)

The authentication service manages users and issues access and refresh tokens.
It provides a CLI to add/remove users and a service to handle authentication request and issue tokens. See [authn service](https://github.com/hiveot/hub/tree/main/authn) for more information.


### Authorization (authz)

The authorization service manages groups that contain consumers and Things.
Consumers that are in the same group as a Thing are authorized to access the Thing based on their role as viewer, operator, manager, administrator or thing. See the [authorization service](https://github.com/hiveot/hub/tree/main/authz) for more information.

### mosquittomgr: Message Bus Manager and Mosquitto auth plugin (deprecated)

Deprecated: This mosquittomgr service will turn into an optional protocol adapter

Interaction with Things takes place via a message bus. [Exposed Things](https://www.w3.org/TR/wot-architecture/#exposed-thing-and-consumed-thing-abstractions) publish their TD document and events onto the bus and subscribe to action messages. Consumers can subscribe to these messages and publish actions to the Thing.

The Mosquitto manager configures the Mosquitto MQTT broker (server) including authentication and authorization of things, services and consumers. See the [mosquittomgr service](https://github.com/hiveot/hub/tree/main/mosquittomgr) for more information.

IoT devices must be able to connect to the message bus through TLS and use client certificate authentication. The Hub library provides protocol bindings to accomplish this.

### directory: Directory Service

The directory service captures TD document publications and lets consumer list and query for known Things. It uses the Authorization service to filter the TD's that a consumer is allowed to see. See the [directory service](https://github.com/hiveot/hub/tree/main/thingdir) for more information.

The directory service is intended for use by consumers. IoT devices only need to use the pub/sub API to publish TDs and events, and subscribe to actions.


## Client Library For Developing IoT Devices And Consumers

Compatible IoT devices must support at least one of the available messaging protocols. The capnp protocol is preferred. Planned alternatives are the MQTT and websocket protocols.

The project provides a [Hub client library for developing IoT devices](https://github.com/hiveot/hub/lib/client) and their consumers. This library provides an implementation of a subset of the [Exposed Thing](https://www.w3.org/TR/wot-scripting-api/#the-exposedthing-interface) and [Consumed Thing](https://www.w3.org/TR/wot-scripting-api/#the-consumedthing-interface) interface with a protocol binding for the messaging. In addition methods to construct WoT compliant Thing Description documents
(TD) are included.

IoT devices will likely also use the [provisioning protocol client](https://github.com/hiveot/hub/idprov/pkg/idprov) to automatically discovery the provisioning server and obtain a certificate used to connect to the message bus.

The above library is written in Golang. Python and Javascript Hub API libraries are planned. They will be added to https://github.com/hiveot/lib/{python}|{js}|{...}
