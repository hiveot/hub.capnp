# Capability Resolver Client

## Introduction
The capability resolver client is intended to easily obtain distributed capabilities regardless of their location or interface protocol.

The use of this resolver client is optional but recommended. The consumer of a capability can remain agnostic of what provides the capability and only only needs to provide the capability's interface type.

A secondary benefit is that it simplifies testing when using an external capability by allowing to register dummy capabilities to force a specific behavior. It is designed for minimal coding overhead, simply get the capability and use it.

Under the hood, this client requires that either a local service is registered that provides the capability, or a marshaller is registered for the RPC protocol that the capability server uses. If the location is known, a URL for a direct connection to the capability over tcp, websocket or unix domain sockets can be provided. If no location is known the resolver service or gateway service is used to obtain as the connection to the capability.  

## Setup

Before using the resolver client, it needs to know the location and protocols to use. 


1. Connect the resolver client to the resolver service: resolver:
```go
 resolver.ConnectToResolverService(url,clientCert,caCert)
```
All parameters are optional. 
* url to the resolver or gateway service. If empty then auto-discovery finds either a locally running resolver or a gateway on the local network.
* clientCert is optional for certificate based authentication. 
* caCert is optional but recommended to ensure the resolver service is legit.

2. Login to the resolver when not using certificate authentication. This step can be skipped if a valid client certificate was provided in step 1.
```go
 resolver.Login(userID, password)
```
3. Register the required capability marshallers for POG (plain old golang) api. Many services have multiple capabilities.
```go
resolver.RegisterCapnpMarshaller[T](factory, url)
```
Where:
* T is the native capability interface type
* factory is a method that takes a capnp client and returns the instance of the marshaller. Each service should include a library with marshallers for the capabilities they provide in the programming languages they support. 

The signature of the marshaller factory depends on the protocol. For example in capnp the directory service marshaller is defined as: 
```
func NewDirectoryCapnpClient(capClient capnp.Client) directory.IDirectory {
```
It takes the capnp RPC client and returns the directory service capability. 

Currently only capnp protocol marshallers are implemented but this approach easily extends to support other protocol marshallers, such as the mqtt protocol. Each protocol will require their specific parameters but all return the native capability interface. 

For example:
```go
resolver.RegisterCapnpMarshaller[directory.IDirectory](capnpclient.NewDirectoryCapnpClient, "")
resolver.RegisterCapnpMarshaller[directory.IReadDirectory](capnpclient.NewReadDirectoryCapnpClient, "")
resolver.RegisterCapnpMarshaller[directory.IUpdateDirectory](capnpclient.NewUpdateDirectoryCapnpClient, "")
````

Alternatively, use RegisterHubMarshallers() which registers all Hub included marshallers.

### Local Capabilities
Locally available capabilities can be registered directly:
```
var service IReadDirectory := MyDirectoryTestStub()
resolver.RegisterService[IReadDirectory](service)
```

This registers a stub service for reading a directory.


### Obtaining a capability

After the application has setup the resolver, using it can't be simpler. To obtain a capability:
```go
 dirCap := resolver.GetCapability[directory.IReadDirectory]()
```

This returns the capability to read the Thing directory. As a convention, capabilities have a 'Release' method that should be called when they are no longer needed to release the resources used.  
