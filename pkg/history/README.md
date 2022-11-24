# Thing Value History 

## Objective

Provide historical reading of thing events and actions.

## Status

This service is functional.

To be completed:
* Enable pubsub subscription for events and actions to update the store when they are published.
* Allow config to determine which store to use


## Summary

History provides capture are restore of past events and actions.

The main use-case is to view and analyze past values of things.  


## Backend Storage

This service uses the bucketstore for the storage backend. The bucketstore supports several implementations including a mongodb time series database. The bucketstore API provides a cursor with key-ranged seek capability which can be used for time-based queries.

All bucket store implementations support this range query through cursors. Since the data volume will be high, the pebble or mongo timeseries stores are preferred choice. 

More testing is needed to determine their limitations.

### Data Size

Data ingress of event samples depends strongly on the type of sensor, actuator or service that captures the data. Below some example cases and the estimated memory to get an idea of the required space.

Since the store uses a bucket per thingID, the thingID itself does not add significant size. The key is the msec timestamp since epoc, approx 15 characters.

Typical samples are around 100 bytes (key:20, event name:10, value: 60, json: 10)

Case 1: sensor with a 1 minute average sample interval. 

* A single sensor -> 500K samples => 50MB/year
* A small system with 10 sensors -> 5M samples => 500MB/year
* A medium system with 100 sensors -> 50M samples => 5GB/year
* A larger system with 1000 sensors -> 500M samples => 50GB/year

Case 2: image snapshot with 10 minute event interval
An image is 720i compressed, around 100K/image. 

* A single image -> 50K snapshots/year => 5 GB/year
* A system with 10 cameras -> 500K snapshots/year => 50 GB/year
* A larger system with 100 cameras -> 5000K snapshots/year => 500 GB/year

Case 3: a timelapse system with a snapshot at the same time each day over a year
Image at 1080 resolution, around 400K/image.

365 snapshots -> 150MB/year
