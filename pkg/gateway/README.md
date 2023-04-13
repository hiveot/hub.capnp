# Gateway Service

The Gateway Service provides the means to obtain capabilities by remote clients and services using the capnproto RPC protocol. This is the primary way to access the Hub.

## Status

This service is functional but breaking changes should be expected.

Features:
* authentication through client certificate
* authentication through login using password
* proxy for capabilities of services that are registered with the hub resolver
* Websocket API with limited capabilities for use by javascript clients

## Summary

The gateway provides network clients the ability to obtain capabilities of Hub services in a secure manner. Services and IoT devices can connect to the Hub and obtain capabilities. Thing devices use publishing of events and subscribing to actions. Users can read the directory, retrieve sensor history, subscribe to events and submit action requests. 

The gateway main purpose is to proxy requests for capabilities to services that provide it. If a call to request a capability cannot be resolved locally, the gateway forwards it to the resolver service, which provides the end point that handles the request.  

The gateway's communication protocol is capnproto.
