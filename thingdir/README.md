# Thing Directory Service 

Golang implementation of a Thing Directory Service client and server library.

## Objective

Provide a shadow registry of a Thing containing Thing Description documents and Thing values.

## Summary

This directory service provides a 'shadow registry' containing Thing Descriptions and Thing values. The service collects TD and event publications on the message bus and stores the information in a TD and a value store. 

The [WoT Directory Specification](https://w3c.github.io/wot-discovery/#exploration-directory-api) describes the requirements for a directory service and is used to guide this implementation. The intent is to be compliant where possible. Note that at the time of implementation this specification is still a draft and subject to change. While the specification covers both service discovery and a directory service, this service focuses exclusively on the directory aspect. For discovery see the '[idprov-go](https://github.com/wostzone/idprov-go)' provisioning plugin.

In WoST, the registration of Thing Descriptions is the responsibility of the Directory Service. Things themselves only have to publish their updates on the message bus without consideration for who uses the information. This separation of concerns has the following benefits:
- allows for centralized access control that is uniform for all Things regardless of their make, model and protocol.
- remove the need for Things to implement authentication and authorization
- further simplifies Thing devices as they only need to publish updates to the message bus
- support bridging between multiple directories without participation by Thing devices 


This package consists of the following parts:

1. Thing Description store for storing and querying Thing Description documents. The current implementation uses a file based store with an in-memory cache. Additional storage backends can be added in the future.

2. Thing event store for storing and querying events. This includes property change events as well as non-property events.

3. MQTT Protocol binding to the WoST message bus. This service subscribes to the message bus and stores published Thing TD documents and events. Things do NOT update the directory themselves, they only need to publish their updates on the message bus.

4. Directory server to serve directory requests. This implements the server side of the directory protocol as described by the WoT Directory API as well as API's to query Thing property values and Thing Events. Authentication is handled using the hub authn service while group and role based authorization is handled through the authz service. The directory server also announces its presence using DNS-SD. 

6. Directory client for golang clients. Intended for clients to query the directory. Clients for other languages will be made available as well, or users can implement their own using the protocol described below.


## Directory API

The directory service supports the API as outlined in the WoT directory specification. 

### Register a Thing TD

Note that Things do not need to use this API if they publish their TD on the message bus.

```http
HTTP PUT https://server:port/things/thingID
{
  ...TD...
}
201 (Created)
```
Other responses:
 * 201 (Update)  - the TD already exists and was replaced
 * 400 (Bad Request) - invalid serialization or TD
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization, or anonymous TDs 

```
A note on anonymous TDs
Anonymous TDs are not allowed in WoST. In order for things to provision and receive a certificate, they must have a thing ID.
```

### Get a Thing TD

```http
HTTP GET https://server:port/things/thingID
200 (OK)
{
  ... TD ...
}
```

Other responses:
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization, or anonymous TDs 
 * 404 (Not Found) - no such thing ID

### Update a Thing TD

Note that Things do not need to use this API if they publish their TD on the message bus.

To replace an existing TD:

```http
HTTP PUT https://server:port/things/thingID
Content-Type: application/td+json 
{
  ...TD...
}
204 (No Content)
```
Other responses:
 * 201 (Created) - Thing didn't exist and was created
 * 400 (Bad Request) - invalid serialization or TD
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization, or anonymous TDs 


To partially update an existing TD:

```http
HTTP PATCH https://server:port/things/thingID
Content-Type: application/merge-patch+json 
{
  ...Partial TD...
}
204 (No Content)
```
Other responses:
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization
 * 404 (Not Found) - TD with the given id not found


### Delete a Thing TD

```http
HTTP DELETE https://server:port/things/thingID
Content-Type: application/td+json 
{
  ...TD...
}
204 (No Content)
```

### Listing of Thing TDs

Example limit nr of results to 10

```http
HTTP GET https://server:port/things?offset=10&limit=10
200 (OK)
Content-Type: application/ld+json
Link: </things?offset=10>; rel="next"
[{TD},...]
```

The optional next link in the response is used to paginate additional results.

Other responses:
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization

### Search For Things With JSONPATH


JSONPATH queries are supported as follows:
> $.td[*].id                   -> list of IDs of things
> $.td[?(@.type=='thetype')]   -> TDs of type 'thetype'
> $.td[0,1]                    -> First two TDs


Example search

```http
HTTP GET https://server:port/things?queryparams=""
200 (OK)
Content-Type: application/json
[{TD},...]
```

Other responses:
 * 400 (Bad Request) - invalid serialization or TD
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization


Where queryparams identify property fields in the TD.

## Property Value API

The property value API provides the last known value of a Thing's properties.
For property value history see the history plugin. 

### Get Property Values Of A Thing

This returns a map of property name-value pairs.

Optional query parameters:
* updatedSince=isodatetime : only return property values that have been updated since the given ISO8601 timestamp.
* propNames=prop1: only return property values for the given names 


```http
HTTP GET https://server:port/values/thingID?[&updatedSince=isodatetime][&propNames=propName1]
200 (OK)
{
  propName1: propValue, 
}
```

### Get Property Values Of Multiple Things

This returns a map of thing ID's with a map of property name-value pairs.

Query parameters must contain a list of ThingIDs:
* things=thingID1,thingID2: List of Things whose properties to get 
* updatedSince=isodatetime : only return property values that have been updated since the given timestamp.
* propNames=propName1,...: only return property values for the given names


```http
HTTP GET https://server:port/values?things=thingID1,thingID2[&updatedSince=isodatetime][&propNames=propname1,...]
200 (OK)
{
  thingID1: {
    propName1: propValue,
    propName2: propValue,
    ... 
  },
}
```


## Security

This service is a WoST Hub plugin and uses the Hub authentication and authorization facilities.

In addition, the following protections are provided:
1. Rate limiting. Limit the number of requests from the same client. [TODO]
2. Request duration. Requests that take too long are aborted. [TODO]
3. Monitoring. Track traffic load and alert on sudden traffic changes. [TODO]

The parameters governing the mitigation can be defined in the service configuration.


## Build and Installation

### System Requirements

This service is core plugin of the WoST Hub. See Hub system requirements for details.

### Manual Installation

See the hub README on plugin installation.


### Build From Source

Build with:
```
make all
```

When successful, the plugin can be found in dist/bin. An example configuration file is provided in config/thingdir.yaml. 

"make install" copies these to the local Hub directory at ~/bin/wost/{bin,config}


## Usage

Configure the service through the config/thingdir-pb.yaml protocol binding configuration file. All settings are optional. The service uses the hub.yaml configuration to determine defaults compatible with the WoST Hub. 

To launch the service simply run dist/bin/thingdir, which subscribes to TDs on the message bus and updates the store. It also launches the service for use by clients to query the directory. 

Currently a file based backend is included. Additional backends can be added in the future.
