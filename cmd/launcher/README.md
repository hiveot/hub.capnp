# Launcher

The Lancher manages the WoST plugins that provide the services to IoT devices, consumers and other services.

## Summary

The launcher is responsible for operating Hub plugins. It purpose is to launch and monitor plugins, and report their status. In itself it does not provide any features other than managing the plugins.

The launcher is configured through the launcher.yaml configuration file that provides a list of services to launch.

The launcher can be built using 'make build', which creates the dist/bin/launcher commandline interface (CLI).

## Usage

See launcher --help 
