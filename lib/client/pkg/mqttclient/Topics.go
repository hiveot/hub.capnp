// Package hubclient with messaging topics for the MQTT protocol binding
package mqttclient

// TopicRoot is the base of the topic
const TopicRoot = "things"

// TopicThingTD topic for thing publishing its TD
const TopicThingTD = TopicRoot + "/{id}/td"

// TopicThingPropertyValues topic for Thing publishing updates to property values
const TopicThingPropertyValues = TopicRoot + "/{id}/values"

// TopicThingEvent topic for thing publishing its Thing events
const TopicThingEvent = TopicRoot + "/{id}/event"

// TopicSetConfig topic request to update Thing configuration attributes
const TopicSetConfig = TopicRoot + "/{id}/config"

// TopicAction topic request to start action
const TopicAction = TopicRoot + "/{id}/action"

// TopicProvisionRequest topic requesting to privision of a thing device
// const TopicProvisionRequest = "provisioning" + "/{id}/request"

// TopicProvisionResponse topic for privisioning of a thing device
// const TopicProvisionResponse = "provisioning" + "/{id}/response"
