# Launcher

The Launcher service manages running of the Hub services. 

## Objectives

The main objective is to manage Hub services and monitor their status. 

## Roadmap

1. Set logging output to log files for each service
2. Support service autostart on startup. Use launcher.yaml config file.
3. Auto restart service if exit with error
4. Restart service if resources (CPU, Memory) exceed configured thresholds
5. Send event when services are started and stopped
6. Send event when resource usage exceeds limits
 

## Summary

The launcher is responsible for starting and stopping services and monitor their operations.

When starting a service, it is launched as a new process. The service keeps the handle on the process and is notified if it terminates.

If a service stops unintentionally it is automatically restarted. If restart fails, a backoff time delays the attempt to start again. This backoff time is slowly increased until a maximum of 1 hour.

To stop a service the launcher simply terminates the process the service runs into and disables its enabled status.

While running, the launcher keeps track of the CPU and memory usage of the service. This is available upon request.


Limitations:
If a service is already running, the launcher does not know about this. Most services might fail if started twice. 


## Launcher Configuration

The launcher uses the following configuration for launching services:
```
{app}/config/launcher.yaml  contains the launcher settings, including the folders to use.
```

See the example file for details.


## Usage

List available services
> launcher list 

Start a service
> launcher start {serviceName}

Stop a service
> launcher stop {serviceName}

Option to use a specific folder with services  
> --services path/to/services
