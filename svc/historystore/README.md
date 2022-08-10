# Thing History Store

The objective of the history store is to provide time based reading of thing property, event and action values.

## Use cases

Main use cases:

1. Show the latest of all PEA (property/event/action) values of a thing (view/edit a thing)
2. View the latest value of a thing-PEA (dashboard panel)
3. View the latest value of multiple thing-PEA (dashboard table)  
   eg: bulk version of 2.
4. View the 24 hours values of a thing PEA for a specific day (graph)
   Plot a dashboard graph or determine min/max/avg
5. View the 24 hours values of multiple thing-PEA of a selected day (multi-line graph)

Bonus:

6. View the values of a thing PEA for a date range (graph)
   Plot a dashboard graph for multiple days
7. View the values of multiple thing PEA for a date range (multi-line graph)
   Plot a dashboard graph for multiple days and sensor values

8. Read a PEA value of a thing at a given date/time (eg look back)
   Example: get temperature value of outdoor multi-sensor at 13:00 of Oct 24th last year
9. Same for multiple thing-PEA values

Extra Bonus:

1. Filter: PEA 'A' of Thing 'T1' less, greater, equal than <value>
   Example: Get indoor temperature T1 when outdoor temperature is below zero
   Example: Get temperature T1 when humidity > 70%
2. Filter: PEA 'A' of Thing 'T1' less, greater, equal than PEA 'B' of Thing 'T2'
   Example: Get indoor temperature T1 when it exceeds outdoor temperature T2
   Example: Get camera snapshot when motion sensor T2 triggered
3. jsonpath filters (as per WoT spec)

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

The use of MongoDB has the added benefit that dapr can be configured to use it as the state store, and ThingStore as well.

The concern with MongoDB is a hefty memory load. Min 256MB and 1GB RAM for 100K assets, although the time series usage significantly more efficient. Another concern is the horrific golang API that can stand in the way to optimize the usage. Write performance is okay with 50K samples/sec.

The Hub's local usage is fairly basic. A small setup with 10 sensors that update every minute would add up to 14K samples a day and 5.3Million samples a year, approx 1GB/year. A large setup with 10K sensors 1TB a year. All reasonable numbers for a small to mid-sized system.

## Data Structure

The history store works with a single data structure of type 'ThingValue'. This type is used to store Events, Properties and Actions:

```
type ThingValue struct {
   ThingID  string `json:"thingID"`
   Name     string `json:"name"`
   Created  string `json:"created"`
   Value    string `json:"value"`
   ValueID  string `json:"valueID"` 
   ActionID string `json:"actionID"`
}
```
ActionID is used to link events to the actions that caused them.  