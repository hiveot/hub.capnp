# Authorization

The Hub manages the authorization of consumers using roles and groups.

## Roles

> Clients have a role that determines what actions they can perform with devices in a group.

Consumers are clients with the role of observer, user or manager. For simplicity the Hub only allows a client to have a single role per group. The type of roles are predefined:

* devices publish Thing information such as TD and events, and subscribe to configuration and actions of things it is the publisher for.
* observers can subscribe to read TD, and Events updates
* users are observers that can publish Thing actions 
* managers are users that can publish Thing configuration 
* superusers are managers that can update group and client roles


## Groups

> Groups bundle Things and Consumers.

A group determines what Things are visible to the consumers of that group based on their role. When a Thing is added to a group, all clients in that group can access the Thing based on their role in the group. When a consumer is added, it is authorized for a role in the group. 

The 'things' group automatically contains all thing. observers, users and admins in the things group can therefore access all things.

For example (work in progress):

> groups.yaml
```yaml
things:
  thing1: ThingRole
  thing2: ThingRole
  thing3: ThingRole

group1:
  user1: ObserverRole 
  user2: UserRole     
  user3: AdminRole    
  thing1: ThingRole   
  thing2: ThingRole   

group2:
  user3: ObserverRole
  thing3: ThingRole
```


