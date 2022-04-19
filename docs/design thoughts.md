# Notes on Hub design.

Just some random notes and thoughts

## Plan

1. Phase 1: Hub MVP [done]
   1. Configuration management for hub and plugins (wostlib-go) [done]
   2. Certificate management (wostlib-go) [done]
   3. MQTT client (wostlib-go) [done]
   4. Launching of plugins (hub) [done]
   5. Logger plugin to track launching problems and test plugins (logger) [done]
   6. Mosquitto no-auth MQTT server configuration (pb_mosquitto) [done]

2. Phase 2: Authentication & Authorization [done]
    1. Certificate based client authentication for plugins and devices [done]
    2. Username/password authentication service for mosquitto/https services [done]
    3. Hub Certificate, Authentication and Authorization CLI [done]

4. Phase 3: Core services [in progress]
    1. Thing provisioning protocol (idprov) [done]
    2. In-memory directory service [done]
    3. Value store (in-memory) [done] 
    4. Web client (thingview) service [in progress] 
   
5. Phase 4: protocol adapters [in progress]
    1.  EDS OWServer 1-wire protocol adapter POC [done]
    2.  Insteon
    3.  IPcam
    4.  Montage
    5.  OpenZwave
    6.  Weather

6. Phase 5: Potential Enhancements
    1. Schema validation [is it worth the hassle?]
    2. Time Series Database for values
    3. Discovery of directory and other services?
    4. Redis persistence backup
    5. Scripting engine plugin
    6. WebSocket protocol adapter
    7. JS Hub Client API's
    8. View discovered things
    9. Dashboard



## Discussion

The hub core consists of:
1. A plugin manager for starting and stopping plugins
1. A configuration manager for loading Hub and plugin configuration
2. An authentication manager for managing plugin, Thing and consumer authentication
3. A pub/sub message bus that relays messages between plugins. This implements the Hub API.

##Protocol Adapters

The Hub provides APIs to provision Things, update Thing Description and properties, send events and handle actions. This API is made available through multiple protocol adapters.

The MQTT protocol adapter manages an external message bus to publish/subscribe to things. Things that connect to the MQTT bus publish using the standard topic schema which is document in the hubapi module. Messages received are routed to the internal message bus.

The HTTP protocol adapters listens for incoming connection requests and passes these on to the internal message bus. 

### Exploring A Single Message Bus Approach 

Use a single message bus for all communication within and to/from the Hub.

1. Needs an message broker. Options:
   1. [Mosquitto](https://github.com/eclipse/mosquitto)
      1. Well known, well tested
      2. light weight that runs out of the box
      3. License: Eclipse Public License 2.0  ???
   2. https://github.com/vardius/message-bus 
      1. License: Apache
   3. https://github.com/nats-io/nats.go   
      1. Easy to use? check
      2. JSON encoding built-in? check
      3. Topic addresses? 
         1. Uses dots '.' separator instead of path. Not MQTT compatible :/
      4. Wildcard subscription? 
         1. Using '*', not MQTT compatible :/
      5. Authentication? JWT
      6. TLS support? Yes, Including self-signed and client certs
      7. License: Apache. Is that a problem?
      8. Integrations with Redis, Apache Spark,, ..., HTTP?, MQTT?
   4. https://github.com/DrmagicE/gmqtt
      1. JSON encoding?
      2. ACLs? Via plugin
         1. Live update of credentials? tbd
         2. Has an HTTP API for that in the admin plugin. Is that good or bad?
            1. Can Hub replace the admin plugin?
      3. Auth methods? password only. Extensible with plugins
      4. TLS support? yes
      5. License: MIT 
      6. Integrations with Redis
2. Authentication rules
   1. Are Plugins restricted?
      1. Why? no reason yet 
   2. Things can pub/sub on their own addresses
      1. Message bus authorizes based on certificate credentials
   3. Consumers
      1. Use roles and group memberships
      2. publish to select things (role or other access control method)
         1. Use group role membership
3. Things can connect to publish updates, events, and subscribe to actions
   1. address is ... 'things/{thingID}/...
4. Consumers can connect to real-time events
   1. address is things/#
5. How do Things that are also consumers identify?
   1. As a Thing. 
   2. Who assigns consumer credentials?
6. Need to protect the topic space again intrusion
   1. Use ACL, deny all and add permissions 

7. There is no difference between thing-plugin, plugin-consumer, thing-consumer communication over the event bus (apart from permissions)
   1. API must be clear on notify of updates from Things vs request to update Things
      1. Option: Use 'set' suffix in topic to indicate request
         1. Consumers are not allowed to publish set

8. Directory service listens for TD messages and services it through the Directory API
   1. API Specification from W3C WoT
   2. MQTT protocol binding not defined so publish TD on things/{thingID}/td

9. How to get property values?
   1. Propery changes are published using events
   2. Event store could be used to collect events
   3. API to query latest and history values
   4. Consumers can subscribe to hub for real-time value updates
   5. Support for readProperties message?

10. Should HTTP API for message bus be supported?
    1. Use-case? 
       1. HTTPS-only clients
          1. N/A. MQTT clients using websocket are widely available
       2. Query TD's and their values without directory service
          1. Makes no sense, might as well use directory service
          2. Sleeping Things might not respond. Directory service doesn't sleep.
       3. Query Thing property/status values -> No
          1. either history service or direct query
          2. sleeping things might not respond
          3. query results would spam everyone on the message bus with value update
          4. potential for DOS attack through message multiplication by one bad actor
          5. increases attack surface by requiring query support on Things 
    2. So NO.  

11. Is there a shadow registry?
    1. What for?
       1. To respond to requests of TDs?
            1. This is the role of the directory service
       2. To service the latest value/event
          1. Option 1: Use the DS
          2. Option 2: Add a separate value store
             1. Would a value store need value schemas from the TD
       3. To service the value/event history 
          1. Option 1: Use the DS
          2. Option 2: Add a separate value store

    2. So ... answer is to use the DS as the shadow registry 
       1. The DS already handles TD queries and authorizes requests. It can do this for values too.
          -> Add API for value query. Follow the HTTP binding for querying Things
       2. Value schema validation can use the already stored TDs
       3. History queries requires an additional time-series store however
          -> Add API to query a value history.

12. Is there schema validation? -> 
    1.  Yes, there should be
    2.  How to validate schema from TD publishers in a performant manner?
        1. ...during Thing provisioning...provisioning doesn't use the message bus
            1.  No. This is not the purpose of provisioning
        2. ...Directory Service validates schema. 
           1. Consumers should use the DS for obtaining TDs, or validate schema themselves

13. Discovery service 
14. listens for plugins on MQTT plugin channel
15. publishes their addresses on DNS-SD

16. Logger service
    1.  Listens on thing publications and writes to files
    2.  Based on configuration

17. Legacy protocol binding connects to Hub API and acts as one or multiple Things
    1.  No provisioning needed 
