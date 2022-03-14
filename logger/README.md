# WoST Hub Logger

This is being changed into a general purpose event logger:
* configurable as to what events to log
  * configuration changes
  * actions
  * adding/removing users
  * password changes
  * auth failure
* configure the topics to log
* include sender ID, timestamp
* content template to remove sensitive data


Simple logger of messages on the hub message bus, intended for testing of plugins and things.


## Objective

Facilitate the development of Things, consumers, and plugins by logging Thing messages on the hub message bus.

## Summary

Things and Hub plugins publish information on Thing TD's, events and actions over the message bus. This plugin writes those messages to file. Each Thing has its own file with the ThingID as its name.

## Build

Use "make all" to build the logger.

The plugin can be found in dist/bin. Copy this to the Hub's bin directory.
An example configuration file is provided in config/logger.yaml. Copy this to the hub config directory.

## Usage

After installation on startup of the Hub. It generates log files in the configured logging folder with the pluginID as the filename. Eg:  {apphome}/logs/{thingID}.log

Instead of logging all Things, the logger.yaml configuration file can be configured with the ID's of the Things to log. The rest will be ignored.
