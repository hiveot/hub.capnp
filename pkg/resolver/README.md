# Hub Resolver

The Hub Resolver locates and connects to capabilities.

A capability is anything expressed through an interface. For example, a directory service capability defines the API of the directory service, the instance or content identifier, and the means to connect to it.

Any HiveOT service can use the Hub Resolver to locate a needed capability and connect to it via capnp without requiring any further domain knowledge. The resolver works with local services, and with remote services using a gateway.

Resolver requirements:
1. The resolver has no hard coded knowledge of the services and their capabilities themselves. As more services and capabilities are added it must be easy to make them available without changes to the resolver.
2. The resolver works with any capnp client.
3. The resolver is not remotely accessible. 
3. Remote clients are supported using a gateway that validates credentials and authorizes access.
4. Remote capabilities can be registered. This is the role of a discovery service.
5. Additional protocols can be added in the future if needed.
   * For example, http access to ip cameras, rtsp access to a video stream, mqtt access to a message bus.
   * Note that only clients that support these protocols can use them.
6. Capabilities that are no longer accessible are automatically removed. As a minimum an hourly renewal interval must be honored.
7. QoS support can be added in future to handle failover and multiple instance services.


Design:

The resolver solution consists of two peers, the client and the resolver service.

Communication between peers uses the two-way capnp protocol. The resolver service is listening on a known Unix or TCP socket address. The client peer connects to that known address.

Each client peer connects to the resolver service and registers its capabilities. The peers maintain this connection for the duration that the capabilities are available. The resolver service therefore maintains a permanent connection with each client peer while running. 

When a client peer connects, it uses the capnp bootstrap facility to provide a peer resolver capability (named CapPeerResolver) to the resolver service,similar to a callback interface. The resolver service uses the client peer capability to obtain a list of available capabilities and to obtain the capability if this is requested by another peer of the resolver service. 

In defining a capability, peers provide the capability name, the protocol used, and the network type (unix or tcp) and connection address. In cases where the protocol is capnp, the capability name is all that is needed.  

The client peer can use the same connection to request another capability from the resolver service.

The resolver can operate in two modes. Direct and indirect:
* In direct mode, the resolver only provides the capability info (ListCapabilities) to the client, which in turn has to connect to the service provider and request the capability using 'GetCapability'. This mode is only available if the service provides a listening socket to connect to.

* In Indirect mode the client requests the capability from the resolver, which obtains it from the service provider through its existing RPC connection. Direct mode can be useful for streaming data while indirect mode allows remote access to capabilities and is more secure as the service does not need to be listening. Indirect mode adds an extra network hop via the resolver service. The capnproto level 3 RPC supports capability handover which removes the extra hop. This is [in development]
  (https://github.com/capnproto/go-capnproto2/issues/160). While services are free to support either or both modes depending on their purpose, it is recommended to at least support the indirect mode. 

The resolver service itself is not remotely reachable. Remote clients connect to a gateway service that handles authentication and authorization of the capabilities, and in turn acts as a proxy of the resolver. The connection to the gateway remains active as long as the offered capabilities remain accessible. The capnp protocol multiplexes all capabilities over the connection. See the gateway service for more information.
