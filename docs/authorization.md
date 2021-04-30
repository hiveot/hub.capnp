# Authorization

The Hub manages the authorization of consumers using roles and groups.

## Roles

> Consumers have roles that determine what actions they can perform

Consumers are clients with the role of observer, user or manager. For simplicity the Hub only allows a client to have a single role. The roles are predefined:

* observers can subscribe to read TD, Events and property updates
* users are observers that can publish actions to Things in the group
* managers can update configuration
* superusers are clients that can update configuration, eg users+manager


## Groups

> Groups bundle Things and Consumers.

A group determines what Things are visible to the consumers of that group based on their role. When a Thing is added to a group, all clients in that group can access the Thing based on their role in the group. When a consumer is added, it is authorized for a role in the group. 

Authorization uses mosquitto ACLs to mimic roles and groups. As the hub writes the authorization configuration into the hub-auth.yaml file, this plugin watches that file for changes and updates mosquitto's ACLs accordingly.


For example (work in progress):

> groups.yaml
```yaml
group1:
  user1: ObserverRole  = read things/.../td|events
  user2: UserRole      = read things/.../td|events write things/.../action
  user3: AdminRole     = readwrite things/.../#
  thing1: ThingRole    = readwrite things/thing1/#
  thing2: ThingRole    = readwrite things/thing2/#

group2:
  user3: ObserverRole
  thing1: ThingRole
```


