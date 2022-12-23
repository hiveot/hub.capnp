# Gateway Service

The Gateway Service provides the means to obtain and register capabilities by remote clients and services using the capnproto RPC protocol.

## Status

This service is functional but breaking changes should be expected.

Planned:
* A middleware chain for:
  - certificate based authentication
  - login name/password authentication
  - logging of requests
  - rate limiting of capabilities based on client type
  - authorization
* HTTP Websocket API for use by web clients


## Summary

The gateway provides network clients the ability to use Hub services in a secure manner. Services and IoT devices can connect to the Hub, publish events, and subscribe to actions. Users can read the directory, retrieve sensor history, subscribe to events and submit action requests. 

The gateway supports the IResolverSession API. Clients use the gateway in the same way as a local resolver would be used with an added authentication 'login' method.

A future consideration - pending on use-cases - is to support registering capabilities to expand the capabilities with remote services, enabling distributed computing through the Hub.  
