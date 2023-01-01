package thing

import "strings"

// MakeThingAddr implements the definition of a Thing's address, eg the publisherID / ThingID
// If publisherID is empty, only the thingID will be included
// This is the de-facto method of constructing a thing's address out of the ThingID's of its publisher and itself
func MakeThingAddr(publisherID string, thingID string) string {
	if publisherID == "" {
		return thingID
	}
	return publisherID + "/" + thingID
}

// SplitThingAddr split the thing address in its publisherID and thingID parts
// If only a single part is found then this is the thingID and publisherID will be empty ""
func SplitThingAddr(thingAddr string) (publisherID string, thingID string) {
	parts := strings.Split(thingAddr, "/")
	if len(parts) == 1 {
		return "", parts[0]
	}
	return parts[0], parts[1]
}

// IsPublisher tests whether the given thingAddr is of the given publisher
// Returns true if it is.
func IsPublisher(publisherID string, thingAddr string) bool {
	parts := strings.Split(thingAddr, "/")
	return publisherID == parts[0]
}
