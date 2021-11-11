// Package td with property value updates
package td

// CreateValuesMessage creates a new values message to report value changes.
// Intended for use by Things.
// Use AddValue to add more values to the message
//  propertyName contains the first property
//  newValue contains the new value
func CreateValueMessage(propertyName string, newValue string) map[string]interface{} {
	values := make(map[string]interface{}, 0)
	values[propertyName] = newValue
	return values
}

// Adds a changed property value to the values message
func AddValue(values map[string]interface{}, propertyName string, newValue string) {
	values[propertyName] = newValue
}
