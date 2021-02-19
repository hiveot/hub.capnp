# Wost Gateway Core Plugins

This folder contains the included core gateway plugins. The purpose of these plugins is to provide essenstial WoST functionality for protocol bindings to WoST compliant devices and services for WoST consumers.

## discovery - WoST Gateway Discovery Protocol Binding

The gateway discovery is a mDNS protocol binding that announces the location of the gateway on the local network. It is intended for Things and consumers to automatically discover the gateway.

## logging - Channel Logging Service

The channel logging service logs channel messages. It is intended as a debugging tool.  

## things - WoST Thing Registration Protocol Binding

The WoST Thing Registrations provides a web API for Things to connect to the gateway. The plugin supports four sub-protocols:
1. provisioning of Things
2. receive TDs from Things
3. receive events from Things
4. send actions to Things

## directory - Directory Service

The directory service provides consumers with a list of known Things.

## admin - Admin Panel

The admin panel service provides an administration panel for viewing and mananging Things through a web browser.

