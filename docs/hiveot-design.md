# Hive Of Things Design

![Design Overview](hiveot-design.png)

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

Even though a distributed architecture might seem overkill for this solution, it opens the door to a resilient decentralized approach where multiple computing units can provide the services without the need of a central coordinating single point of failure. Any computing device can participate to run services, increasing the capacity of the system with little or no dedicated hardware. 

## Infrastructure

The Hub infrastructure is build on [Cap'n Proto](https://capnproto.org/) and leverages [capabily based security](https://en.wikipedia.org/wiki/Capability-based_security). Cap'n proto can use various types of transport, like a simple pipe, Unix domain sockets, and TCP sockets. Communication is encrypted using TLS.

Internally, Hub services communicate using capnp over Unix Domain sockets. Externally, the Hub Gateways provide capnp APIs using TCP sockets.  

A gateway services provides capabilities to the services that are available on the Hub itself or via the gateway of a participating device. While the extra hop to a different gateway might add extra latency, the (planned) use of [level 3 protocol](https://capnproto.org/rpc.html) will allow the capability to be transferred using a direct connect to the offering service. Level 3 is in development for the go-capnp implementation in 2022.

If identical capabilities are available from multiple gateways then the connected gateway automatically selects the most efficient option. This allows for automatic failover in case a gateway is no longer available.

Access to capabilities via non-capability protocols, such as HTTP, MQTT, gRPC, is provided through so-called protocol gateway services. These services provide a bridge between the external facing protocol and the internal capnp protocol. In order to access these services authorization is required, based on the gateway protocol. For example BASIC, DIGEST, or OAUTH2 authentication.

While a Gateway provides the ability to access capabilities, the available capabilities are constrained by authorization of the client. Client authorization can be obtained by using a client certificate or login credentials for authentication.

To determine which capability an authenticated client has access to, the gateway uses the authorization service. The rules are based on the type of client, eg IoT device, service, or user, and the role in groups the client shares with Things.


## Gateway

The Hub includes a gateway service to facilitate communication between the Hub and the world around it. The Gateway is intended for interaction with IoT devices, services and end-users.

Most protocol bindings and end users will use the gateway to interact with the Hub. The gateway offers a hub capabilities API that is based on the type of authenticated client: IoT devices, services including protocol bindings, and end users. This is security feature reduces to attack footprint of the API as only allowed capabilities are available.

The built-in capabilities are provided by core services. Capabilities can be expanded by  service that registers new capabilities through the gateway. These services can reside on the Hub computing device or on any external device. 

The communication protocol used by the gateway and core services is 'capnproto'. The capabilities concept is at the heart of this protocol. Addition protocol bindings are supported such as websockets and mqtt (planned) for specific tasks. 

This following paragraphs describes the capabilities for various types of clients the gateway caters to.

### IoT Devices

Most IoT devices contain sensors or actuators. The Hub handles these based on the W3C WoT (Web of Things) standard that defines how IoT devices are described. The Gateway service provides capabilities for IoT devices to:

1. Provision IoT devices with a client certificate
2. Publish IoT device [WoT TD documents](README-TD.md)
3. Publish IoT device events as described in the device TD
4. Subscribe to requested actions as described in the device TD

IoT devices can use the gateway provided capabilities using the capnp protocol. Publish and subscribe capabilities are also available via websockets.

New IoT devices must be 'provisioned' before they can publish their information. The gateway offers the provisioning capability using an out-of-bound secret. On successful provisioning the device receives a signed authentication certificate. This certificate identifies the device and enables capabilities to publish a TD document, corresponding events and to subscribe to actions. IoT device certificates are relatively short lived and can be renewed before they expire. 

### IoT Protocol Bindings

To use devices that implement their own protocol, a protocol binding service is needed. 

IoT protocol bindings are services that translate between a 3rd party IoT protocol and the Gateway API. For example, the zwave protocol binding translates between ZWave node device information and WoT (Web of Things) standardized documents, events and actions. 
New bindings are added on an ongoing basis. For more information see the [bindings repository](https://github.com/hiveot/bindings)

IoT Protocol Bindings receive the same capabilities as IoT devices. The main difference is is that they tend to publish for multiple devices.

### End-Users

End-users are people that use a client application to connect to the Hub, like a web client or a mobile client.

These clients connect to the gateway and authenticate using their credentials, usually a login ID and a password. 

To end-users the gateway offers the capability to:
1. Authenticate the end-user using login ID and password.
2. Retrieve available Things from the directory. 
3. Read historical values from the history store.
4. Subscribe to Thing events.
5. Publish Thing actions.

To unauthenticated users only the login capability is available. 

In addition to password based authentication, the gateway also allows the use of client certificates, which is useful especially for machine to machine authentication. 

### Services

Services have access to most of the Hub capabilities including device capabilities and end-user capabilities. They can be publishers of information, like a weather service, or consumers of information like automation rules. 

Services authenticate using a service certificate that is issued by the administrator. The certificate identifies the service and is not transferable.  


## Core Capabilities

Core services provide the capabilities needed to be able to function as a Hub. These services are not required to all live on the same computing devices and can be distributed among several. 

### Pub/Sub Communication

Consumers subscribe to events and can publish actions, while IoT devices or services publish events and subscribe to actions. The internal pubsub service uses capnproto streams for publish subscribe.

The pubsub service itself offers a simple publish and subscribe API. Clients of the API get respective capabilities that are constrained based on the type of client:
- all clients must provide authentication in order to obtain pub/sub capabilities 
- end-users are constrained to groups they are a member of
- IoT bindings are constrained to publishing events and subscribing to actions for their publisher only.
- Services can publish events of which they are the publisher.
 
Binding services MUST handle authentication and authorization of the client.

The pub/sub follows the WoT model of events and actions. TD documents and events always require a publisher ID which is the ID used to authenticate with. 

### Provisioning

Provisioning is the process of pairing an IoT device to the Hub. 

IoT devices that support the [idprov protocol](https://github.com/hiveot/idprov-standard) can automatically discover the Hub on the local network using the DNS-SD protocol and initiate the provisioning process. When accepted, a CA signed device (client) certificate is issued.

The device certificate supports machine to machine authentication between IoT device and Hub. See [idprov service](https://github.com/hiveot/hub/tree/main/pkg/provisioning) for more information.

### Authentication (authn)

The authentication service manages end-users, and issues access and refresh tokens. See [authn service](https://github.com/hiveot/hub/tree/main/pkg/authn) for more information.


### Authorization (authz)

The authorization service manages groups that contain consumers and Things.
Consumers that are in the same group as a Thing are authorized to access the Thing based on their role as viewer, operator, manager, administrator or thing. See the [authorization service](https://github.com/hiveot/hub/tree/main/pkg/authz) for more information.


### Directory 

The directory service captures TD document publications and lets consumer list and query for known Things. It uses the Authorization service to filter the TD's that a consumer is allowed to see. See the [directory service](https://github.com/hiveot/hub/tree/main/thingdir) for more information.

The directory service is intended for use by consumers. IoT devices only need to use the pub/sub API to publish TDs and events, and subscribe to actions.


### History 

The history service provides capabilities to query for Thing event values in the past. Like the directory service it is intended for use by consumers. IoT devices only need to use the pub/sub API to publish events.

As some events publishers can be quite noisy and generate a lot of useless data, it is possible to blacklist events from certain publishers.

The service can be replaced with a more advanced implementation if needed.

