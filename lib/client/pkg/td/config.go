// Package td with configuration update request message
package td

// CreateConfigRequest creates a new message for requesting updates to configuration properties
// Use AddConfigRequest to add more configuration properties to the message
//  propertyName contains the first configuration property to request
//  newValue contains the new requested value
//
// This returns a message that can be published with IHubClient.PublishConfigRequest()
func CreateConfigRequest(propertyName string, newValue string) map[string]interface{} {
	config := make(map[string]interface{}, 0)
	config[propertyName] = newValue
	return config
}

// Add a property value to a configuration request
func AddConfigRequest(config map[string]interface{}, propertyName string, newValue string) {
	config[propertyName] = newValue
}
