# State Storage

The state storage provides a simple key-value storage services to store application state.

## Objective

Persist the state of other services and consumers.

## Summary

Services that need to persist state can use this key-value store to save their state. This is also available to consumers that like to store state of their web application, for example a dashboard layout.

The default state store uses the built-in in-memory Key-value store with a file based storage. Additional storage engines can be made available in time.

The state service is intended for a relatively small amount of data. Limits for the total data size can be set for services and consumers. The default is 100 keys with 100K values for a total of 10MB per service or user. See state.yaml for the selection of backend storage and limits.

The state service is intended to be accessed via the gateway. The gateway determines the location of the service and provides the capability to access the service at  that location. 

By default, the state store resides on the Hub server.

## Usage

The service is intended to be started by the launcher. For testing purposes a manual startup is also possible. In this case the configuration file can be specified using the -c commandline option.

The service API is defined with capnproto IDL at:
> github.com/hiveot/hub.capnp/hubapi/state.capnp

A goland interface can be found at:
> github.com/hiveot/hub/pkg/state/IStateStore
 
### Golang POGS Client

An easy to use golang POGS client can be found at:
> github.com/hiveot/hub/pkg/state/capnpclient

 
To use the client, first obtain the capability (see below) and create the POGS client. For example to write a state value:

```golang
  stateCap := GetCapability(appID) // from authorized source
  stateAPI := NewStateCapnpClient(stateCap)
  stateAPI.Put("mykey", "myvalue")
```

where GetCapability provides the capability to read and write key-values. This is restricted to the user's ID and application.

The capability can be obtained from the gateway service with proper authentication.
