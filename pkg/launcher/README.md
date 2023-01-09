# Launcher

## Objectives

The main objective is to manage Hub services and protocol bindings and monitor their status. 

## Features

Completed:
1. List, Start and Stop available services and protocol bindings
2. Set logging output to log files for each service
3. Support service autostart on startup. Use launcher.yaml config file.
4. Track service/binding memory and CPU usage. 

Planned:
4. Auto restart service if exit with error
4. Restart service if resources (CPU, Memory) exceed configured thresholds
5. Send event when services are started and stopped
6. Send event when resource usage exceeds limits
 

## Summary

The launcher is responsible for starting and stopping services and protocol bindings, and monitor their running status.

When starting a service or binding, it is launched as a new process. Services and bindings terminate on the SIGTERM signal.

If a service stops unintentionally it is automatically restarted. If restart fails, a backoff time delays the attempt to start again. This backoff time is slowly increased until a maximum of 1 hour.

To stop a service the launcher simply terminates the process the service runs into and disables its enabled status.

While running, the launcher keeps track of the CPU and memory usage of the service. This is available upon request.

**Limitations:**

* The launcher will not recognize services started on their own. Services will not function properly when started twice.


## Launcher Configuration

The launcher uses the following configuration for launching services:
```
{app}/config/launcher.yaml  contains the launcher settings, including the folders to use.
```

See the example file for details.


## Usage

List available services and bindings
```sh
launcher list 
```
Start a service or binding, or all services/bindings
```sh
launcher start {serviceName} | all
```

Stop a service/binding
```sh
launcher stop {serviceName} | all
```
