# mauth - Mosquitto Auth Golang Plugin.

This package is used to generate a shared object (.so) file that is a mosquitto plugin. It invokes authorization for subscribing and publishing of topics.

The mosquitto configuration 'mosquitto.conf' - generated from mosquitto.conf.template - links to this .so file.  

Generating this file for ARM should be done on a compatible machine. Setting up a cross compilation environment in x86 is too much work at this point.


Dependencies:
* libssl-dev
* mosquitto-dev