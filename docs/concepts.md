# Secured Things Gateway (STG)

The STG is a gateway for Secured Things. It primary purpose is to provide access to Secured Things for authorized consumers.

The STG core is implemented in Golang to keep it small and without runtime dependencies. Gateway and related functionality is provided through plugins. The core consists of a workflow pipeline that plugins publish onto and receive from via Websockets.

## Overview 
The STG core is a workflow pipeline engine that uses plugins to collect, store, process and consume data. This is modelled based on [Omar Elgabry's article in medium.com](https://medium.com/omarelgabrys-blog/data-pipelines-8bc38c89f501). 

In case of WoST, plugins can be written in any language, including ECMAScript to be compliant with the [WoT Scripting API](https://www.w3.org/TR/wot-architecture/#sec-scripting-api). 

Different types of information are processed differently and thus have their own pipeline. The pipeline configuration defines the pipeline ID and data schema. Schemas are based on JSON-LD as used in schema.org, WoT and NGSI-LD.

Data in a pipeline goes through 3 sequential stages: Capture, process and consume. Each stage has its own plugins to provide functionality for that stage. 

Capture plugins capture data from 3rd party sources and push it into the pipeline identified by a pipeline ID. For example, a TD (Thing Description) registration plugin can receive TD's and push them into the pipeline with the ID of 'TD'. Multiple capture protocols can be supported for example capturing a TD through MQTT. Pipeline IDs and schemas are standardized but plugins do have the ability to define their own pipeline IDs and schemas.

Once raw data is pushed into the pipeline, it is stored before processing. By default this store is a simple in-memory object store that persists the data until the final consume step has completed. It can be replaced with a full blown database if there is a need to archive and replay the raw data. 

Processing plugins register themselves to receive data from a pipeline. When multiple processing plugins are registered for the same pipeline they each are invoked sequentially in order of registration. Processing plugins can enrich the data or push it into a new pipeline. For example an image recognition plugin can identify objects in an image. An alerting plugin can watch for an event, check if a condition is met and push an event into the alert pipeline. If a plugin returns a failure the pipeline is aborted. A plugin can simply return the provided data which is passed along the pipeline.

After data processing plugins are invoked, the pipeline moves on to consumer plugins.

Consumer plugins listen on their pipeline and are invoked with the result of the processing stage. They take this data and consume the content. For example, an email plugin can listen on the alerting pipeline and email alerts based on their severity. 
The history consumer provides storage of all data processed by the workflow. It has a query API for other consumers to query past data. The default is a short term in-memory store which can be swapper by persistent storage.


## Configuration

The pipeline engine is configured with an optional YAML file that lets the user change the default settings.

```json
{
   listen: string,     // Listening address and port. Default is localhost:9678
   plugins: string,    // Folder with plugins. Default is {stg folder}/plugins
}
```

## Launching Plugins

Plugins are launched at startup and given three arguments: 
* {listen} containing the IP and port where the core listens for websocket connections.
* {authorization} containing the authorization token the plugin must include when establishing its websocket connection.
* {configFile} containing the path to the plugin YAML configuration file. This is optional. If possible plugins should function out of the box without configuration.

## Plugin Connection

After launch plugins connect to their pipeline socket. The address is made up as follows:
> https://{listen}/pipeline/{ID}/{stage}

Where:
* {listen} is the address and port of the websocket server. The default is localhost:9678 for internal connections. When plugins reside on other systems this has to be the gateway IP address that can be reached by that system.
* pipeline is the keyword for all pipeline addresses
* {ID} is the ID of the pipeline. This is associated with a schema for the pipeline data format that the plugin must be able to read.
* {stage} is the pipeline stage, one of: "capture", "process", "consume"

A valid authorization header token must be present. This token is provided on startup. The core will reject any connection requests that do not contain a valid token.

While a plugin can make as many connections as needed it is strongly recommended to adhere to the single responsibility principle and only connect to the pipeline and stage that is needed to fulfil that responsibility. 

## Capture Plugins

The primary role of capture plugins is to capture data from 3rd party sources and push it into its pipeline. The format of the data pushed into the pipeline MUST match the schema associated with the pipeline ID. Schemas are defined in the JSON-LD format as defined in schema.org, WoT schemas, and NGSI-LD schemas. 

For example, the schema for the [Thing Description](https://www.w3.org/TR/wot-thing-description/#behavior-data) is descripted in the [TD Schema](https://www.w3.org/TR/wot-thing-description/#json-schema-for-validation)

Capture plugins connect to the pipeline capture stage on address:
> https://{listen}/pipeline/{ID}/capture

## Processing Plugins

The primary role of processing plugins is to validate, enrich, or transform the data in the pipeline. If the data is changed, it can be pushed into a new pipeline, similar to a capture plugin.

Processing plugins provide an Invoke method that is called with newly stored data. The method returns the data to be forwarded to the next plugin. It can also act as a circuit breaker to abort further processing and end the pipeline.  

Capture plugins listen on the pipeline process stage address:
> https://{listen}/pipeline/{ID}/process


The plugin MUST send an appropriate result, which is one of:
1. The same message. Processing will continue.
2. A modified message of the same type. Processing will continue with the modified message. For example, an annotated image.
3. A null message that acts as a circuit breaker. Processing will end.


## Consumer Plugins

The primary role of consumer plugins is to take the data after processing and make it available to consumers. This can mean sending it as an email, or storing it in a database for later retrieval.  

Consumer plugins listen on the pipeline consume stage address:
> https://{listen}/pipeline/{ID}/consume

The plugin can decide based on the content of the message whether to consume the message or ignore it. No response is expected.


## Writing Plugins

Plugins can be written in any programming language. They can include a configuration file that describes their purpose and the pipeline they use. Plugins must use websockets to publish and listen on pipelines. 

There is nearly no boilerplate code involved in writing plugins except for making a websocket connection and listening for messages. Plugins can therefore be very lightweight and efficient. 

Usually plugins run in their own process, isolated from other plugins. It is however possible to write a plugin that launches other plugins in threads. For example, a JS plugin can load additional plugins written in Javascript. Each of the additional plugin must connect using the websocket for their pipeline stage.

# Using Secured Things Gateway

To use the gateway, the appropriate plugins for the intended purpose must be selected. 

The capture plugins that should be enabled are:
* Discovery capture plugin [mDNS]
* Provisioning capture plugin
* TD capture plugin
* Directory capture plugin

The recommended process plugins are:
* Scripting engine to run custom scripts [ECMA/Javascript, Python]

The recommended consumer plugins are:
* Intermediary plugin to connect to a cloud gateway

Depending on the purpose additional plugins can be enabled:
* Intermediary plugin 
* Web server administration console plugin 
* History plugin [SQLite, MongoDB, Time Series DB]
* Notification plugin [Email, SMS, Slack, ...]

