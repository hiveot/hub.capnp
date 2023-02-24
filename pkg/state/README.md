# State Service

The state service provides a simple key-value storage service to store application state.

## Objective

Persist the state of other services and consumers.


## Summary

Services that need to persist state can use this key-value store to save their state. This is also available to consumers that like to store state of their web application, for example a dashboard layout.

The state service uses the bucket store package that supports multiple implementations of the key-value store. The default store uses the built-in btree key-value store with a file based storage.  

The state service is intended for a relatively small amount of state data. Performance and memory consumption are good for at least 100K total records. 

## Usage

The service is intended to be started by the launcher. For testing purposes a manual startup is also possible. In this case the configuration file can be specified using the -c commandline option.

The service API is defined with capnproto IDL at:
> github.com/hiveot/hub/api/hubapi/state.capnp

A goland interface can be found at:
> github.com/hiveot/hub/pkg/state/IStateStore
 
### Golang POGS Client

An easy to use golang POGS client can be found at:
> github.com/hiveot/hub/pkg/state/capnpclient

 
To use the client, first obtain the capability (see below) and create the POGS client. For example to write a state value:

```golang  (not full code)
  stateCap := GetCapability(appID) // from authorized source
  stateAPI := NewStateCapnpClient(stateCap)
  bucket := stateAPI.CapClientBucket(id)
  bucket.Put("mykey", "myvalue")
```

where GetCapability provides the capability to use storage buckets for the client to read and write key-values. 
