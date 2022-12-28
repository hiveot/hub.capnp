# Hub Resolver

The Hub Resolver locates and connects to capabilities provided by other services.

## Objective

Allow any Hub service to obtain capabilities from other services by name, regardless the location of the service.


## Resolver Requirements

1. The resolver has no hard coded knowledge of the services and their capabilities themselves. As more services and capabilities are added it must be easy to make them available without changes to the resolver.
2. The resolver works with any capnp client.
3. The resolver is not directly remotely accessible. 
3. Remote clients are supported using a gateway that validates credentials and authorizes access.
4. Remote capabilities can be registered. This is the role of a discovery service.
5. Capabilities that are no longer accessible are automatically removed. 
6. QoS support can be added in future to handle failover and multiple instance services.


## Design

The resolver consists of a service and a capability provider. 

The resolver service collects a list of capabilities available through the services. To this end it monitors the socket folder and connects to each of the services using their socket. If the service supports the ListCapability interface, its capabilities are made available through the resolver.

To make ListCapabilities available, a service's capnp server includes the use of the 'CapServer'. CapServer implements the ListCapabilities method and provides info on capabilities that are exported by the service. 

The resolver can operate in two modes. Direct and indirect:
* In direct mode, the resolver only provides the capability info (ListCapabilities) to the client, which in turn connects to the service provider directly and request the capability.

* In Indirect mode the client requests the capability from the resolver, which forwards the request to the service that supports it. When a request for a method that doesn't exist comes in, the resolver determines the service that does have the method and forwards the request. The result is passed back to the caller.

Note that getting a capability through the resolver adds approx 40% to the duration of calls of that capability, as traffic flows through the resolver to the service. This can be avoided using the connection information from ListCapabilities a direct connection to the service can be made that does not suffer this overhead.

The resolver service itself is not remotely reachable. Remote clients connect to a gateway service that handles authentication and authorization of the capabilities. The gateway uses the resolver to get the list of capabilities and connects directly to the services.
