// Package thing with methods to handle Thing IDs
package thing

import (
	"fmt"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
	"strings"
)

// CreatePublisherID creates a globally unique Thing ID that includes the zone and publisher
// name where the Thing originates from. The publisher is especially useful where protocol
// bindings create thing IDs. In this case the publisher is the gateway used by the protocol binding
// or the PB itself.  See also SplitThingID.
//
// This creates a Thing ID: URN:zone:publisher:deviceID:deviceType
//  zone is the name of the zone the device is part of. Use "" for local.
//  publisher is the deviceID of the publisher of the thing.
//  deviceID is the ID of the device to use as part of the Thing ID
func CreatePublisherID(zone string, publisher string, deviceID string, deviceType vocab.DeviceType) string {
	thingID := fmt.Sprintf("urn:%s:%s:%s:%s", zone, publisher, deviceID, deviceType)
	return thingID
}

// CreateThingID creates a ThingID from the zone it belongs to, the hardware device ID and device Type
// This creates a Thing ID: URN:zone:deviceID:deviceType.
//  zone is the name of the zone the device is part of. Use "" for local.
//  deviceID is the ID of the device to use as part of the Thing ID.
func CreateThingID(zone string, deviceID string, deviceType vocab.DeviceType) string {
	if zone == "" {
		zone = "local"
	}
	thingID := fmt.Sprintf("urn:%s:%s:%s", zone, deviceID, deviceType)
	return thingID
}

// SplitThingID takes a ThingID and breaks it down into individual parts.
// Supported formats:
//  A full thingID with zone and publisher: URN:zone:publisherID:deviceID:deviceType.
//  A thingID without publisher: URN:zone:deviceID:deviceType
//  A thingID without zone: URN:deviceID:deviceType
//  A thingID without anything specific: URN:deviceID
func SplitThingID(thingID string) (
	zone string, publisherID string, deviceID string, deviceType vocab.DeviceType) {
	parts := strings.Split(thingID, ":")
	if len(parts) < 2 || strings.ToLower(parts[0]) != "urn" {
		// not a valid thing ID.
		// Handle graceful by using the whole ID as deviceID
		return "", "", parts[0], ""
	} else if len(parts) == 5 {
		// this is a full thingID  zone:publisher:deviceID:deviceType
		zone = parts[1]
		publisherID = parts[2]
		deviceID = parts[3]
		deviceType = vocab.DeviceType(parts[4])
	} else if len(parts) == 4 {
		// this is a partial thingID  zone:deviceID:deviceType
		zone = parts[1]
		deviceID = parts[2]
		deviceType = vocab.DeviceType(parts[3])
	} else if len(parts) == 3 {
		// this is a partial thingID  deviceID:deviceType
		deviceID = parts[1]
		deviceType = vocab.DeviceType(parts[2])
	} else if len(parts) == 2 {
		// the thingID is the deviceID
		deviceID = parts[1]
	}
	return
}
