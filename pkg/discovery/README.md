# discovery 

## Introduction 

The discovery package provides discovery of Hub services.

Services that depend on other services need a means to obtain the capability to that service. 

Examples:
* History and Directory use the pubsub service to receive events and TD documents
* Provisioning service uses certs service
* Any service can use the state service to persist their state
* The gateway service uses the provisioning and pubsub services 

Rather than hard-coding service addresses, discovery provides a mechanism to determine how to obtain the capability of services by name. 

This decouples the location of the services. As long as discovery can find the service, other services can use it. Services can therefore operate decentralized on any device, including available IoT devices, as long as they can be discovered and connected to.

## Roadmap

Support for discovery is planned in stages:

### Stage 1: local discovery

At stage 1 all services live on the same device. The discovery service uses the local gateway service to obtain a list of available capabilities.


### Stage 2: remote discovery

Stage 2 includes services that live on other devices.

The discovery service builds a list of available capabilities both local and remote using the information gathered from the gateway service.

Devices running the discovery service optionally publish their presence on the local network using DNS-SD. When a another discovery service is found, a connection is made to its gateway and its list of available services is obtained. Periodically the process repeats and updates take place.  

Alternatively, the discovery service is configured with a list of devices that run a gateway service. In this case the DNS-SD protocol does not need to be used. This can also be used in case of multiple subnets, such as VPN, where DNS-SD is blocked between networks.

Clients of the discovery service can search for a capability and receive the best location of the gateway providing this capability, if available. The client then proceeds to connect to the provided gateway service and request its available capabilities. The discovery service act basically as a cache of remote gateways and their services. 

Once the client determines the location of a capability, it connects to the gateway and requests an instance of the capability.. The client then continues to use it as if it was local.

Note that automatic discovery requires that devices that make Hub services available must also run the discovery service as well. 

### Stage 3: failover

Stage 3 allows multiple identical services on the local network. If one fails, another will be used, effectively creating a failover.

Discovery periodically scans the network for services and ranks them based on their QOS rating.
When a service is requested, it offers the most efficient service. 

The QOS rating is periodically updated. Each discovery service tracks the cpu and memory usage of services and provides a rank based on available resources. In addition, the discovery client measure the connection latency to determine the final rating. 


#### Stage 4: single instance failover

Some services can only have a single instance active on the network. In this case an election algorithm that uses each service ranking chooses which instance is active and announces this to the other running discovery services. Once all services are in sync, they apply the new rank. 

As HiveOT is decentralized, no central decision maker is present. Discovery services make a 'hive decision' in which the majority rules.


#### Stage 5: Service Hive

The discovery service can work in conjunction with the launcher service. Rather than having the same service running as standby on multiple devices, only one is running, freeing up the resources on the other device. If a new instance is needed to balance the load or as failover, the discovery services can elect a new device to run the service and start a new instance using the launcher and make it available.

Services can be short or long lived. The launcher can kill a service if the device runs low on resources which will lower its rating. The discovery service will share the new rating with other services and will elect a replacement based on the new ratings. 

This mechanism allows devices that are used for multiple tasks to be used optimally. For example, an ML process that is only used intermittently can provide other services in its spare time. Once an ML tasks is running these services are terminated. When the ML task is ended they can be  made available again.

#### Stage 6: The Internet Hive  

External resources can be pulled in by allowing additional cloud resource to be used. Cloud resources can be included through use of a VPN or through a Bridge Service.

When a VPN is used, the additional devices will automatically be discovery, as long as the DNS-SD is tunneled over the VPN. If DNS-SD is not available, a hard-coded discovery configuration can provide the address of other services.

When the Bridge Service is used, two Hubs contact each other in a peer-to-peer connection. This connection allows to 1) share Thing information, and 2) make device resources/service available through tunneling the discovery information.

If Stage 6 is needed then the Hub has reached full maturity. This stage is only useful if there is a need to share services across the internet with trusted 3rd parties. One example is to include a cloud service provider for additional resources. Note that the bridge service alone can already share IoT data between two locations, so this does not require the Discovery service. 
