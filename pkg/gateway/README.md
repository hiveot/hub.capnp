# Capnp Gateway 

## Introduction
The Capnp Gateway provides the means to connect to Hub services using the capnproto RPC protocol.

Features:
1. Client authentication using certificates. TLS is required.
2. Client authentication using username/password. TLS is required.
3. Capabilities directory.
4. Issuing of capabilities.
5. Authorization when obtaining capability.

Usage:
The gateway is intended to be used by any user of the Hub such as IoT devices, services, and end-users. These can access the Hub's services by requesting a desired capability from the gateway. The gateway authenticates the client, authorizes it, and obtains and returns the capability.

Local services also use the gateway to obtain a capability by name. The difference is that the gateway client can be asked to directly connect to local capabilities if available, instead of using the gateway service as an intermediary. This minimizes latency. If the capability isn't available locally it will be obtained through the service.

Use of the included gateway client (in golang at least) is therefore preferred over using the service via a raw capnp protocol interface. The client can bypass the service if requested. It will try a direct connection using UDS to the service based on the service name and retrieve the capability from that service. If named pipes or memory-mapped communication is available, it can be used as well. Anything for speed :)

The capnp protocol is platform agnostics so this would work similar in C++ and javascript as it does in golang. The plan is to add gateway clients for those languages as well.   

## Status

This service is in development.



## Registering Capabilities

There is a need to register service capabilities with the gateway. The gateway has no knowledge of the capabilities provided by services. In addition, many capabilities are obtained indirectly by invoking a method on the service. Last, as more capabilities are added, it must be easy to make them available via the gateway without updating the gateway.

Option 1

Exported capabilities are registered in the service configuration file at config/{serviceName}.yaml

Capabilities available through the gateway are listed in the capabilities section of its configuration file along with the client type they are intended for.

```yaml
# Exported capabilities 
capabilities:
  # this method is available to IoT devices, services, end-users and event unauthenticated users
  method1:
    clientType: 
    - iotdevice
    - service
    - user
    - noauth

  #
  method2 :
    clientType:
      - iotdevice
```


Alt:
Each service implements IHiveService which has the methods: 
* getCapabilities which returns a list of available capabilities and client type.
* getCapability returns the capability provided by the service
  * this uses the capability type name which the service converts to the implementation
* problem: how to apply constraints, such as limiting the capability to the clientID
* pro: no config needed
* pro: can obtain derived capability directly 


## How to use - under investigation

IBIS: (issue based discussion)

1. What capabilities can local clients obtain?
   * Since local clients are always services, they have full access to other services (for now)
   * Also, using UDS, how to use cert for auth?
2. What can remote clients obtain? 
   * Remote clients can only obtain capabilities intended for their role: service, iotdevice, end-user.
   * End-users have an additional role wrt Things in the groups they are a member of.
3. Who approves capability requests? -> middleware
   * Middleware hooked into the gateway service handles authorization, logging and other tasks
   ? how to log each requests? or should it hook into the service? -> pogs generated wrapper ?
4. How are capability requests handled from capnp client to capnp service?
   1. Client creates a gateway client instance for capnp and provides local/remote connection info (UDS or TCP)
      * a fixed local UDS address or TCP address provided through discovery 
   2. Client authenticates with the gateway, providing credentials, including client type and client ID 
      * gateway service stores credentials in the client session for further requests
   3. Gateway capnp server receives capability request and passes it to the POGS service
      * method is getCapabilityCapnp and returns a capnp api
      * each protocol binding will need its own protocol api 
   4. POGS gateway service connects to capnp service and receives service capabilities (IService api)
      * services only expose the listCapabilities and getCapability APIs which are defined in the IService API
        * reason: allow gateway to authorize getting the capability 
      * all other capabilities must be obtained first using getCapability, which is subject to auth.
   5. Gateway requests service capability using the standard IService API
   6. Service returns capability to gateway
   7. Gateway returns capability to client
   8. Client converts capability instance to actual type
5. How are capability requests issued using WebSockets?
   * Client creates a websocket client and connects to the websocket gateway
   * Gateway verifies authentication level of the client 
   * Client requests capability
   * Gateway receives request through middleware, which does auth
   * Gateway issues request to service using capnp
   * Service returns capability
   * Gateway stores capability in the websocket adapter context for further use and returns an ack
   * Client invokes a method on the capability
   * Adapter invokes the capnp capability - how to identify interface and method id?
     * service provides it through standard IService API -> list capabilities -> ids+names
6. How are capability requests approved? Middleware
   * middleware also handles logging, tracing, authentication, rate limiting, resiliency (crash handler)
7. Where is middleware hooked in? 
   * A: on the capnp RPC side as this works for all cases...
     * how does the middle receive authentication info?
       * context?
     * how do local services authenticate? how to determine the clientID?  
     * Except for authentication which is handled by the protocol adapter
   * B: on the protocol adapter side.
8. How to measure performance of each request? latency, logging, rate limiting?
   * capnp doesn't support this and there are no hooks :(
   * option 1: modify compiler to provide hooks
   * option 2: include in client/server POGS wrappers - eventually include in generated POGS code.
     * client wrapper adds timestamp to requests
     * server wrapper does rate limiting and logging and uses client timestamp to measure latency
