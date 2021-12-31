# Thing View

This thingview plugin lets users view and update Things through a webbrowser.

## Objective

Provide an user interface to view, control and configure Things using a web bowser.

## Status

The status of this plugin is in-development.

## Audience

This project is aimed at IoT developers that value the security and interoperability that WoST brings. WoST Things are more secure than traditional IoT devices as they do not run a server, but instead connect to a Hub to publish their information and receive actions.

## Summary

This plugin consists of a web client that presents a user interface to display, control and configure Things. This client is implemented in Vue-3 and Typescript.

This client uses the account service to login and manage its configuration, including user configuration for the dashboard. After login it uses the Thing Directory Service API to retrieve discovered Things and connects to the MQTT message bus to receive status updates from Thing devices and send action and configuration update to Thing devices. 

A valid user login is required to use this service. A cookie holds the authentication refresh token so login is only needed after it has expired, typically a couple of weeks of no usage.

Authentication is provided by the Hub auth service. This service supports a user configuration key-object map to store the client's configuration. This enables the web client to retain its dashboard configuration across multiple devices.




## Build and Installation


### Dependencies (tentative)

* Language Framework: Vuejs + Typescript

* Build tool: Vite

Vite is simple, fast and just works, even with VSCode debugging. 
The omission of webpack avoids wasting hours on the build and debug problems. 

For UI component library Element-plus is used, which works out of the box.

Reactive storage is built using Vue 3 built-in reactivity features, whic removes the need for reactive storage such as VueX and cuts down on a large amount of boilerplate code.

* Other Dependencies
  + VueUse - the kitchensink of vue 3


### Build (tentative)

Run make all to build. This will download VUE-3 and dependencies from npm.

## Usage

Point a webbrowser to the server {hostname}:443 dependent on the configuration.
A login page prompts. The Hub auth utility can be used by the administrator to create a login. 

By default the client assumes that the Hub server lives on the same address as the node server for the webclient. The account setup supports a different remote auth and directory server address.



