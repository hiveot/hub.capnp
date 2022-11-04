# Hub Usage 

Short description of usage and bindings where HiveOT can be useful.

## IoT Data Collection
1. ZWave binding
2. OWServer binding
3. ISY99x Insteon binding
4. IPCam binding. Snapshots from IP cameras.
5. SNMP network binding. Identify network devices and status.
7. Philips HUE binding
8. Zigbee binding

## External services integration
1. Openweathermap binding
   * Obtain weather forecasts and warnings by location; Weather radar.
2. Canada weather forecast binding
   * Obtain weather forecasts and warnings by location; Weather radar.
3. pihole integration. 
   * Identify destinations using DNS outgoing connections. 
   * Determine privacy concerns.
   * Easy configuration of pihole via hiveot; Block certain destinations.
4. Android GPS location service.
   * User location tracking
4. Email integration for notifications
5. SMS integration for notifications
6. Bridge to share information with other Hubs
7. Twitter integration
   * listen for keywords on official channels, eg HWY 1 closed.
8. Denon AVR control: ON,OFF,Channel,Volume
9. Roku integration, watch channel and select show. 
10. Search engine integration for querying data
    * Sonic: https://github.com/valeriansaliou/sonic

## Analysis & Aggregation

1. IPCam snapshot with augmented info such as temperature, history, ..
2. Wallpaper combines multiple IPCam snapshots into a screen background.
3. Security camera motion detection
4. Image identification using ML
5. Voice command recognition using ML
6. Security analysis using motion detection and IP cameras

## Notifications Outputs

1. Send email 
2. Send SMS message
3. Publish twitter message
4. 'hiveot' mobile app
5. TV output (display alert) 

## Automation

1. ITTT type automation, If 'event' and 'condition' then 'action'
2. Auto shutoff water valve after X minutes
3. Auto on/off front door lights based on time of day and sunset time 
4. If motion detected, trigger PTZ camera to preset location and record for X minutes
5. When pumphouse/wet area temperature drops below threshold send an alert.



# OLD



Short description of generalized use-cases the Hub aims to serve along with the use of capabilities.

## IoT Device Provisioning

Before IoT devices and external services can be used, they must be 'provisioned' or registered with the Hub. Provisioning establishes a trust between Hub and the IoT device using certificates which grants the device the capability to publish events and receive actions with the Hub.

In the provisioning process, the IoT device identity is verified and the Hub issues a client certificate. There can be multiple methods of provisioning, manual and automated.

With the manual provisioning method, the administrator uses the device's public key to create a certificate using the Hub CLI, and installs that certificate on the IoT device. If the device does not have a public key, the Hub creates a private/public key pair for it that must be installed on the device along with the certificate. The certificate is signed by the Hub's CA and recognized by the Hub.

The Hub also supports the '[idprov](http://github.com/hiveot/idprov-standard)' provisioning protocol. The 'idprov' protocol supports both capnp and https transports. It uses an out-of-band provided secret to automatically provision the device. The administrator can upload a list of device-IDs and corresponding device secret to the Hub using the CLI or web client. Once the device is activated, it will try to provision using these two parameters. If they match then the certificate is issued.

For security reasons 'idprov' issued device certificates are only valid for 30 days. The device has to renew the certificate before this expires using the idprov renewal capability, provided during provisioning. 

If the certificate is no longer valid then the request will be stored and await administrator approval. The device will periodically retry and once the administrator approves, it will receive a new certificate. Only then it regains the capability to publish events and receive actions.   


3rd party protocols such as zwave, zigbee, 1-wire, coap, are supported through protocol binding services, or 'IoT gateway service'. An IoT gateway service connects to the 3rd party device or gateway, creates a TD document for each discovered 'Thing' and publishes this TD with the Hub. This follows the exact same process as an IoT device. Hence, IoT services must therefor also provision with the Hub to obtain and renew a certificate.

An IoT gateway service can also be a bridge to another Hub that shares select 'Things'. The bridge is the publisher for Things from the other Hub. This follows the same process as IoT devices and IoT services.

## IoT Devices Publish 'Thing Description' (TD) Documents 

The Hub keeps a directory of 'Things' that have been published by IoT devices and services. A 'Thing' is described using the [W3C WoT Thing Description](https://www.w3.org/TR/wot-thing-description11/) standard.

Provisioning gives the IoT device or service the capability to publish events and receive actions for one or more Things controlled by this device. After provisioning is complete, it publishes events containing a 'TD' document for each of the Things it controls. The Hub stores the TD in its directory for consumers to discover and use. 

Provisioning has provided the IoT device the capability to publish events and receive actions for Things of which it is the publisher. This leverages the 'capabilites' aspect of capnp and provides implicit security by restricting activities to either publish events or receive actions for Things of which it is the publisher.

IoT devices or services also publish a TD that describes themselves in addition to the sensors, actuators or other Things it controls. This enables configuration of the device or service.

## IoT Devices Publish Events 

The TD document describes, amongst others, Thing attributes and events.  

When a Thing value changes, its IoT device/service publishes an event with the Hub. An event can be anything that is described in its TD events or properties section, such as property value change, sensor value change, actuator state change, a service output change and so on.

To send an event, an IoT device or service uses the capability received during provisioning. This capability is a capnp RPC method constrained to publishing of events for the IoT device.

Question: when an IoT device is disconnected and reconnected, is this capability still valid or does it need to be re-acquired?
Question: How can capnp limit the capability to just publishing for one particular publisher. Does the publisher need to be specified or is that part of the received capability?

## IoT Devices Receive Actions For Things

The TD document describes, amongst others, Things, its attributes and actions. Attributes are used to present Thing information and to configure it.

When an attribute value is writable, it is a thing configuration. This configuration change and any action described in its TD can be triggered by sending an action request to the IoT device. 

Actions are 'requests' and do not have to be accepted by the Thing. Once they are applied however they result in an event that describes that change. It is good practice for the IoT device to also send an event when the request is rejected. This event can link back to the action that caused it so consumers can receive confirmation of the action.

The IoT device has received the capability to receive actions for the Things it is the publisher. On startup it listens for actions requested from the Hub. The capnp protocol protects against forging actions so only actions from services that have received the capability will be received. Security is implicit and no additional security checks are needed by the device. 

Question: Can an IoT device re-use a capability to receive actions between restarts or must it retrieve a new one?


## Consumers Retrieve TD Documents

Before consumers can use IoT devices they need to know which ones are available and what their capabilities are. This is described in TD documents that are stored in the directory. 

The consumer only has access to TD documents of Things that are in the same group as the consumer. When receiving the list of TD documents from the Hub, the list of available Things is therefore obtained from the groups the consumer is a member of. This list is then used to obtain the TD documents from the directory. 

After login, the consumer receives the capability to retrieve this list of TD documents, and the capability to control them based on the user role.
The groups are managed separately by the administrator.


## Consumers Retrieve A Thing Property Values

Once a list of available TD's is received, consumers can retrieve the property values of a thing.

Question: normally this happens using the Thing-ID. What is the best approach with capnp? Does capnp provide a capability for each Thing separately? 

For realtime updates, the consumer subscribes to events from a Thing. The capability to subscribe is provided with the received list available Things.  

Question: Is the above idiomatic for capnp?
