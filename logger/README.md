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


## Status 

The status of this plugin is Alpha. Basic logging of Thing messages is functional but breaking changes must be expected.


## Audience

This project is aimed at software developers, system implementors and people with a keen interest in the Web of Things. 

## Summary

Things and Hub plugins publish information on Thing TD's, events and actions over the message bus. This plugin writes those messages to file. Each thing has its own file to enable testing the output of each Thing.

## Build and Installation

### System Requirements

This plugin runs as a plugin of the WoST hub. It has no additional requirements other than a working hub message bus.


### Manual Installation

See the [Hub README](https://github.com/wostzone/hub/blob/main/README.md#plugin-installation) on plugin installation.

In short: copy this plugin to the Hub bin folder and the logger.yaml config file to the Hub's config folder. Add the logger module to the list of plugins to launch on startup and restart the Hub.


### Build From Source

Build with:
```
make all
```

The plugin can be found in dist/bin. Copy this to the Hub's bin directory.
An example configuration file is provided in config/logger.yaml. Copy this to the hub config directory.

## Usage

When installed as a Hub plugin this plugin is launched automatically on startup of the Hub. It generates log files in the configured logging folder with the pluginID as the filename. Eg:  {apphome}/logs/{thingID}.log

Instead of logging all Things, the logger.yaml configuration file can be configured with the ID's of the Things to log. The rest will be ignored.
