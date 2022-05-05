# Thing History Service

## Objective

Provide time series access to historical Thing property and event values using pluggable storage backends.

## Summary

This service collect Thing event data and pushes values into a time series database for further analysis and visualization. A query API allows consumers to use historical data into dashboard graphs. This API support authentication and authorization support using the Hub's role based groups.

The backend storage of thing data is provided through 'adapters'. The adapters provide access to a larger backend storage such as time series databases like OpenTSDB, TDEngine and VictoriaMetrics. A bare-bone in-memory store is included for testing and simple use-cases.

Time series databases can be used to aggregate data from one or multiple Hubs to provide small to large scale analytics and visualization. Support for authentication and authorization for the time series database is out of scope for the WoST Hub. 

This package consists of the following parts:

1. Service for querying time series data with support for authentication and authorization. This implements the ThingData API and is intended for use by Hub consumers.  

2. MQTT protocol binding to the WoST message bus for collecting and storing Thing events and property values.

3. Storage adapters for various backends:

   3.1. Basic in-memory store for storing and querying events and property values in time buckets. This implementation uses in-memory maps for lookup and periodically writes to disk. Memory consumption can be controlled through the configuration of buckets.  -  tbd, is there an existing library for this? 
   
   3.2. Adapter for writing/querying data with VictoriaMetrics time series database.  [todo]
   VictoriaMetrics allows for large scale data collection and analysis, supports Grafana and Prometheus for visualization.

   3.3. Adapter for writing/querying data with LinDB [todo]

   3.4. Adapter for writing/querying data with OpenTSDB [todo]

   3.5. Adapter for writing/querying data with TDEngine [todo]


## ThingData API

API to store and query Thing Data.

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

### Get Event History Of A Thing

This returns the events sent by a Thing since the given start time.

Optional query parameters
* start=iso8601 datetime. Default is ?
* end=iso8601 datetime. Default is 'now'
* eventNames=list of event names. Default is all events.


```http
HTTP GET https://server:port/events/thingID?[&eventNames=eventname1,...][&start=isodatetime][&end=isodatetime]
200 (OK)
{
  timestamp: [{event}, ...]
  or
  eventName: [{event}, ...],
  ...
}
```
### Get Event History Of Multiple Things

This returns the events sent by Things since the given start time.

* things=thingID1,thingID2: List of Things whose events to get. Default is all.
* eventNames=list of event names. Default is all events.
* start=iso8601 datetime. Default is 1 hour ago
* end=iso8601 datetime. Default is 'now'

```http
HTTP GET https://server:port/events?[things=thingID1,thingID2][&eventNames=eventname1,...][&start=isodatetime][&end=isodatetime]
200 (OK)
{
  thingID1: {
    eventName: [{event}],
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

Depending on the chosen storage adapter, additional services and memory can be required.

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
