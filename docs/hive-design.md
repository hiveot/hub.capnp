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

### Hive Gateway

The Hive Gateway is the entry point to the capabilities of services in the Hive. Each Hive computing device runs a gateway that acts as a proxy to its services. Direct access to services and IoT devices is blocked. Only gateway access is possible.

Hive gateways communicate using the capnp protocol as their purpose is to provide capabilities, which is unique to capnp.

Hive gateways can discover each other through DNS-SD and offer capabilities from these discovered gateways. 

While the extra hop to a different gateway might add extra latency, the (planned) use of [level 3 protocol](https://capnproto.org/rpc.html) will allow the capability to be transferred using a direct connect to the offering service. Level 3 is in development for the go-capnp implementation as of mid 2022. 

If identical capabilities are available from multiple gateways then the connected gateway automatically selects the most efficient option. This allows for automatic failover in case a gateway is no longer available.

Access to capabilities via non-capability protocols, such as HTTP, MQTT, gRPC, is provided through so-called protocol bindings, which are services running on a Hive device. These service provide a bridge between the external facing protocol and the internal capnp protocol. In order to access these services a authorization token is required, which is provided by an (planned) authorization protocol binding using BASIC, DIGIST, or OAUTH2.

While the Gateway provides the ability to obtain capabilities, the available capabilities are constrained by authorization of the client. Client authorization can be obtained by using a client certificate or login credentials for authentication. 

To determine which capability an authorized user has access to, the gateway uses the authorization service. The rules are based on the type of client, eg IoT device, service, consumer, and its role, viewer, operator or administrator. Not all capabilities are immediately available, as some can require additional end-user input.

Some examples of capability constraints:
1. IoT devices can only use the publish/subscribe capability for Things of which they are the publisher.  
2. Consumers can only use the subscribe/read/query capabilities for Things that are in the same group as they are.
3. Consumers can only use the publish/write capability for Things that are in the group they have the 'operator' or 'admin' role.   

These constraints are embedded in the capability that is provided by the service. The gateway is unaware of these constraints and simply passes the capability on to the client.


### IoT Protocol Binding Communication

Communication with IoT devices or services take place through IoT protocol binding services. PB services connect using the 3rd party IoT device protocol and converts this to Hub messages with WoT prescribed documents. This enables the Hub to communicate with any IoT device or service for which a protocol is available.

Protocol bindings are responsible for:
* Create a 'TD' Thing Description document for available Things
* Send events when Thing properties or outputs change value.
* Pass on actions requested via the Hub to the Thing's IoT device

### IoT Device Communication

IoT devices or services that are HiveOT compatible can directly communicate with the Hub gateway without the need of for a protocol binding. The gateway provides a set of services to work with these devices/services.

1. Provisioning process
    1. The administrator provides a list of pre-approved devices and their secrets
    2. An IoT device can submit a provisioning request with or without an OOB (out of band) secret
    3. The administrator can view a list of provisioning requests for approval
    4. The administrator can approve a request
    5. An IoT device receives provisioning approval along with a signed certificate used for secure machine-to-machine communication.
2. IoT device publishes one or more TDs of Things available through the device
3. IoT device publishes an event when Thing properties or output values change
4. IoT device receives a request for action

There is little difference between IoT device communication and protocol binding communication to the Hub. They both perform the actions described above. 

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

### Authentication

The authentication service manages users and issues access and refresh tokens.
It provides a CLI to add/remove users and a service to handle authentication request and issue tokens. See [authn service](https://github.com/hiveot/hub/tree/main/authn) for more information.


### Authorization

The authorization service manages groups that contain consumers and Things.
Consumers that are in the same group as a Thing are authorized to access the Thing based on their role as viewer, operator, manager, administrator or thing. See the [authorization service](https://github.com/hiveot/hub/tree/main/authz) for more information.

### mosquittomgr: Message Bus Manager and Mosquitto auth plugin (deprecated)

Deprecated: This mosquittomgr service will turn into an optional protocol adapter

Interaction with Things takes place via a message bus. [Exposed Things](https://www.w3.org/TR/wot-architecture/#exposed-thing-and-consumed-thing-abstractions) publish their TD document and events onto the bus and subscribe to action messages. Consumers can subscribe to these messages and publish actions to the Thing.

The Mosquitto manager configures the Mosquitto MQTT broker (server) including authentication and authorization of things, services and consumers. See the [mosquittomgr service](https://github.com/hiveot/hub/tree/main/mosquittomgr) for more information.

IoT devices must be able to connect to the message bus through TLS and use client certificate authentication. The Hub library provides protocol bindings to accomplish this.

### thingdir: Directory Service

The directory service captures TD document publications and lets consumer list and query for known Things. It uses the Authorization service to filter the TD's that a consumer is allowed to see. See the [directory service](https://github.com/hiveot/hub/tree/main/thingdir) for more information.

The directory service is intended for use by consumers. IoT devices only need to use the pub/sub API to publish TDs and events, and subscribe to actions.


## Client Library For Developing IoT Devices And Consumers

Compatible IoT devices must support at least one of the available messaging protocols. The capnp protocol is preferred. Planned alternatives are the MQTT and websocket protocols.

The project provides a [Hub client library for developing IoT devices](https://github.com/hiveot/hub/lib/client) and their consumers. This library provides an implementation of a subset of the [Exposed Thing](https://www.w3.org/TR/wot-scripting-api/#the-exposedthing-interface) and [Consumed Thing](https://www.w3.org/TR/wot-scripting-api/#the-consumedthing-interface) interface with a protocol binding for the messaging. In addition methods to construct WoT compliant Thing Description documents
(TD) are included.

IoT devices will likely also use the [provisioning protocol client](https://github.com/hiveot/hub/idprov/pkg/idprov) to automatically discovery the provisioning server and obtain a certificate used to connect to the message bus.

The above library is written in Golang. Python and Javascript Hub API libraries are planned. They will be added to https://github.com/hiveot/lib/{python}|{js}|{...}
