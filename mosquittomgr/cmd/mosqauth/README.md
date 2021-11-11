# mosqauth - Mosquitto Auth Golang Plugin.

## Introduction

This package generates a shared object file (mosqauth.so) that is a mosquitto plugin. It invokes authentication and group based authorization for access to topics.

The mosquitto configuration 'mosquitto.conf' - generated from mosquitto.conf.template - links to this .so file.  

Generating this file for ARM should be done on a compatible machine. Setting up a cross compilation environment in x86 is too much work at this point.

Dependencies:
* libssl-dev
* mosquitto-dev


## Usage

There is nothing to do other than to install mosquittomgr using the make file. Mosquitto will invoke the plugin authentication and authorization handlers as needed.


## Notes

This plugin uses the file based authentication and authorization stores. 

A future improvement could be to use an auth service instead.