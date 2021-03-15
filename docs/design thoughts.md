# Notes on Hub design.

## Plan

1. Phase 1: Hub MVP
   1. Configuration management for hub and plugins (api-go)
   2. Certificate management, create CA, Server, Client certs (hub)
   3. Launching of plugins (hub)
   4. Logger plugin to track launching problems and test plugins (logger)
   5. Mosquitto no-aut MQTT server configuration (pb_mosquitto)
   6. HTTP protocol adapter (pb_http)
   7. MQTT client (api-go)
   8. HTTP client (api-go)
   9.  DNS-SD local discovery protocol binding (pb_discovery)
   10. In-memory directory service
   11. EDS OWServer 1-wire protocol adapter

2.  Phase 2: Authentication & provisioning
    1. Add username/password authentication to mosquitto
    2. Add Basic security schema to HTTP
    3. Add Digest security schema to HTTP
    4. Add Bearer security schema to HTTP
    5. Add PKS security schema to HTTP
    6. Thing provisioning protocol (hub+pb_mosquitto)
    7. Provisioning approval process

3.  Phase 3: Directory service
    1.  In-memory directory service
    2.  Schema validation
   
4.  Phase 4: Add more protocol adapters
    1.  Insteon
    2.  IPcam
    3.  Montage
    4.  OpenZwave
    5.  Weather

5.  Phase 5: Potential Enhancements
    1.  Hub CLI 
    2.  Redis persistence backup
    3.  Scripting engine plugin
    4.  WebSocket protocol adapter
    5.  JS Hub Client API's
    6.  View discovered things
    8.  Dashboard



## Discussion

The core consists of:
1. A plugin manager for starting and stopping plugins
1. A configuration manager for loading Hub and plugin configuration
2. An authentication manager for managing plugin, Thing and consumer authentication
3. A pub/sub message bus that relays messages between plugins. This implements the Hub API.

##Protocol Adapters

The Hub provides APIs to provision Things, update Thing Description and properties, send events and handle actions. This API is made available through multiple protocol adapters.

The MQTT protocol adapter connects to an external message bus to publish/subscribe to things. Things that connect to the MQTT bus publish using the standard topic schema which the adapter subscribes to. Messages received are routed to the internal message bus. If the internal message bus is the MQTT bus then there is nothing to do.

The WebSocket protocol adapter listens for websocket connections on the Hub API addresses. Messages received are send to the internal message bus. 

The HTTP protocol adapters listens for incoming connection requests and passes these on to the internal message bus. 

What is the internal message bus made of?
A: MQTT bus. Convenient for existing MQTT users.
B: WebSocket message bus. Runs out of the box

### Exploring A Single Message Bus Approach 

Use a single message bus for all commonication within and to/from the Hub.

1. Needs an message broker  
   1. use Mosquitto for external use-cases
   2. find/create a websocket pub/sub client/server for internal 
      1. why reinvent the wheel?
         1. Needs something light weight that runs out of the box
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
   1. Plugins are unrestricted 
      1. Plugin manager creates Mosquitto credentials for plugins
   2. Things can pub/sub on their own addresses
      1. Provisioning process (manager?) creates credentials and ACL
   3. Consumers
      1. Use roles and group memberships?
      2. publish to select things (role or other access control method)
         1. Use group role membership
3. Things can connect to publish updates, events, and subscribe to actions
   1. address is 'things/{id}/...
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

8. Directory service listens for TD and property updates and servces it through the Directory API

9.  HTTP API routes request onto the message bus. 
   1. PUT request can go directly onto the bus
   2. GET requests read from service cache

10. Is there a shadow registry?
    1. What for?
       1.  To respond to requests of TDs and Thing values?
            1.  Is that the Directory Service?
                1. Maybe both? What is the difference?
    2.  When not...
        1. How to get TDs by their ID without Directory Service?
           1. WoT specifies to query the Thing which we cannot do. What is Hub's equivalent?
                1. A: publish request, wait for response by ... some service
                2. B: Implement cache in the HTTP and WebSocket APIs 
           2. You don't. If you know the ID you almost always already know the TD...
        2. How does an intermediary service share TD's?
           1. Its own cache
    3. So ... answer is No. 
       1. The directory service acts as a shadow registry
       2. The DS can also respond to message bus requests 

11. Is there schema validation? -> 
    1.  Yes, there should be
    2.  How to validate schema from TD publishers in a performant manner?
        1.  ...during Thing provisioning...provisioning doesn't use the message bus
            1.  No. This is not the purpose of provisioning
        2.  ...In parallel. Allow the publication but log schema validations
            1.  Pro: allow direct access to the message bus
            2.  Pro: Invalid but usable schemas can still be used
            3.  Con: consumers might see invalid schemas and have to be resilient 

12. Discovery service 
  1. listens for plugins on MQTT plugin channel
  2. publishes their addresses on DNS-SD

13. Logger service
    1.  Listens on thing publications and writes to files
    2.  Based on configuration

14. Legacy protocol binding connects to Hub API and acts as one or multiple Things
    1.  No provisioning needed 


