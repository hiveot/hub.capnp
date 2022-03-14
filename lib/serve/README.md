# WoST Hub Server Library

This repository provides a library with definitions and methods to provide services as part of the WoST Hub. For developing clients of the Hub see 'hubclient-go'.

## Summary

This library provides functions for creating WoST Hub services. Developers can use it to create a secure TLS server that:
- manage certificates for clients
- works with the Hub authentication mechanism using client certificate authentication and username/password authentication. 
- provides DNS-SD discovery of the service
- watch configuration files for changes

A Python and Javascript version is planned for the future.
See also the [WoST Hub client library](github.com/wostzone/hubclient-go) for additional client oriented features. 

### discovery

### hubnet

Helper functions that a server might need. Obtain bearer token for authentication, determine the outbound interface.  

### tlsserver

Server of HTTP/TLS connections that supports certificate and username/password authentication, and authorization.

Used by the IDProv protocol server and the Thingdir directory server.

### watcher

File watcher  
