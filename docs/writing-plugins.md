# Writing Hub Plugins

This section provides instructions on how to write a plugin for use with the WoST Hub. 


> ## Under Development
> This document is under development



## Starting And Stopping Plugins

Plugins are launched by the Hub and passed the same commandline parameters used to start the Hub itself. These commandline parameters can be used to determine the location of configuration files, certificates and log files.

Commandline options:
```
-c file.yaml           Use this configuration file instead of the default config/hub.yaml
 
--home=path            Set the application working directory. Default is parent of the executable 

--certFolder=path      Set a different TLS certificate folder. Default is ./certs

--configFolder=path    Set a different config folder. Default is ./config

--hostname=dnsname|ip  Message bus address host:port". Default is localhost:9678

--logFile=path         Write log message to this file. Default is ./logs/{pluginID}.log

--protocol=name        Message bus protocol: internal|mqtt

--pluginFolder=path    Alternate plugin folder. Empty to not load plugins.

--logLevel=level       Set loglevel to one of: error|warning|info|debug. Default is warning
```

Upon startup a plugin should read the commandline to determine the configuration file and read the hub.yaml and plugin configuration. A library to do this in a single line is included in the "hubconfig" package of the hubapi-go repository:

```golang
  import "github.com/wostzone/hubapi/pkg/hubconfig"

  pluginConfig := MyPluginConfig{}
  hubConfig, err := config.SetupConfig(homeFolder, pluginName, &pluginConfig) 
```

Next, initialize the plugin code, connect to the Hub message bus and wait for the SIGTERM signal:

```golang
	hub.WaitForSignal()
```


## Connecting To The Hub Message Bus

After loading the configuration, plugins connect to the message bus as a thing and/or consumer. The message bus address and port are configured in the config/hub.yaml configuration file.

Using the Hub configuration connect to the message bus:

```golang
  import "github.com/wostzone/hubapi/pkg/wostmqtt"

  client := wostmqtt.NewThinglient(hostname, port, certFolder, clientID, credentials)
  err := client.Start()
```

The default connect address for the MQTT message bus:
> mqtt://localhost:8883/

Where localhost and port are defined in the Hub messaging configuration. The plugin reads the hub.yaml on startup to ensure the same messaging setup is used. 

The provided client libraries in hubapi-go implement the connection logic for the various protocols so the plugin developer only needs to apply the configuration of the address and port. Next, the plugin can use the Client API to receive, send or update TDs, send and receive events and action messages.


## Stopping The Plugin

When the Hub terminates it stops each plugin in turn using a SIGTERM signal.
The plugin should close connections and exit with result code 0.


