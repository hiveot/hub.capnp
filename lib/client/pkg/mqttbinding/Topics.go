// Package mqttbinding with messaging topics for the MQTT protocol binding
package mqttbinding

// TopicThingTD topic for thing publishing its TD
const TopicThingTD = "things/{id}/thing"

// TopicThingEvent root topic for thing publishing its Thing events
const TopicThingEvent = "things/{id}/event"

// TopicAction root topic request to start action
const TopicAction = "things/{id}/action"

// TopicThingProperty root topic for publishing property value updates
const TopicThingProperty = "things/{id}/property"

// TopicProvisionRequest topic requesting to provision of a thing device
// const TopicProvisionRequest = "provisioning" + "/{id}/request"

// TopicProvisionResponse topic for provisioning of a thing device
// const TopicProvisionResponse = "provisioning" + "/{id}/response"
