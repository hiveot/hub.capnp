# Thing Event History Store

The objective of the history store is to provide time-based reading of thing events, properties and actions.

## Use cases

Main use cases:

1. Show the latest of all of a thing values of a thing (view a thing)
2. View the latest of a single thing value with a 24 hour min, max and average (dashboard panel) 
3. View the latest of select values of one or more things (dashboard table)  
4. View the value over a 24 hours period with (graph)

Bonus:

5. View the value of a thing for a date range (graph)
   Plot a dashboard graph for multiple days

6. Read a nearest value of a thing at a given date/time (eg look back)
   Example: get temperature value of outdoor multi-sensor at 13:00 of Oct 24th last year

Extra Bonus:

7. Filter: Get property X of a Thing when property Y is less, greater, equal than <value>
   Example: Get indoor temperature T1 when outdoor temperature is below zero
   Example: Get temperature T1 when humidity > 70%
8. Filter: get property X of Thing when it is less, greater or equal than property Y of another Thing
   Example: Get indoor temperature T1 of Thing A when it exceeds outdoor temperature T2 of Thing B
   Example: Get camera snapshot when motion sensor T2 triggered
9. jsonpath filters (as per WoT spec)

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

The use of MongoDB has the added benefit that it can be used as the state store, and ThingStore as well. 

The concern with MongoDB is a hefty memory load. Min 256MB and 1GB RAM for 100K assets, although the time series usage significantly more efficient. Another concern is the horrific golang API that can stand in the way to optimize the usage. Write performance is okay with 50K samples/sec. Last, 

The Hub's local usage is fairly basic. A small setup with 10 sensors that update every minute would add 14K samples a day and 5.3Million samples a year, approx 1GB/year. A large setup with 10K sensors 1TB a year. All reasonable numbers for a small to mid-sized system.

## Data Structure

The history store works with a single data structure of type 'ThingValue'. This type is used to store Events, Properties and Actions:

```
type ThingValue struct {
   ThingID  string `json:"thingID"`
   Name     string `json:"name"`
   Created  string `json:"created"`
   Value    string `json:"value"`
}
```
