# Thing Value History 

## Objective

Provide historical reading of thing events and actions.

## Summary

History provides capture are restore of past events and actions.

The main use-case is to view and analyze past values of things.  

## Roadmap

1. Use a secure connection to the mongodb database.
2. TBD determine what other query features are needed, if any.
3. TBD Add support for SQLite

## Backend Storage

The history backend is best served with a time-series database that can handle:

* open-source
* easy to setup and maintain
* minimal memory requirement of < 100MB, 1 CPU
* low maintenance, schemaless
* ingress of up to 100 samples/sec (1000 sensors @ 10 second interval, 10M/day, 3B/year)
* small document size, approx 300 characters/sample
* long storage period, 5 years and up, given enough disk space
* data import and export
* time-to-live for short term high resolution data storage.
* query support
   * downsampling. Viewing time series over a year with a few hundred points at the most.
   * filter on json data fields

Nice to have:

* integration with reporting tools, grafana, prometheus, ?
* SQL query support for further integration
* Geo area query/filter
* Additional query languages such as R, ...
* cold storage option
* use with dapr
* golang adapter

Database candidates that match these requirements are InfluxDB, MongoDB, QuestDB, VictoriaMetrics. Embedded tsdb's that are considered are BuntDB or NutsDB.

The use of MongoDB has the added benefit that it can be used as the state store, and Directory store as well. 

The concern with MongoDB is a hefty memory load. Min 256MB and 1GB RAM for 100K assets, although the time series usage significantly more efficient. Another concern is the horrific golang API that can stand in the way to optimize the usage. Write performance is okay with 50K samples/sec.  

The Hub's local usage is fairly basic. A small setup with 10 sensors that update every minute would add 14K samples a day and 5.3Million samples a year, approx 1GB/year. A large setup with 10K sensors 1TB a year. All reasonable numbers for a small to mid-sized system.

Alternative storage engines will be considered for the future, including SQLite.

## Data Structure

The history store works with a single data structure of type 'ThingValue'. This type is used to store Events and Actions:

```
type ThingValue struct {
   ThingID    string `json:"thingID"`    // ID of the thing whose value is stored
   Name       string `json:"name"`       // Name of the event or action whose value is stored
   Created    string `json:"created"`    // Timestamp the value was created in ISO8601 format (YYYY-MM-DDTHH:MM:SS.sss-TZ)
   ValueJSON  string `json:"valueJSON"`  // The JSON encoded value
}
```
