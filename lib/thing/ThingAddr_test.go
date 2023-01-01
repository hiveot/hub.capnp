package thing_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/hiveot/hub/lib/thing"
)

func TestIsPublisher(t *testing.T) {
	const testDevice1 = "device1ID"
	const testDevice2 = "device2ID"
	logrus.Infof("---TestIsPublisher---")
	thingAddr1 := thing.MakeThingAddr(testDevice1, "urn:sensor1:temperature")
	thingAddr2 := thing.MakeThingAddr(testDevice2, "urn:device2:sensor1")
	thingAddr3 := thing.MakeThingAddr(testDevice2, "")

	// setup
	isPublisher := thing.IsPublisher(testDevice1, thingAddr1)
	assert.True(t, isPublisher)
	isPublisher = thing.IsPublisher(testDevice1, thingAddr2)
	assert.False(t, isPublisher)
	isPublisher = thing.IsPublisher(testDevice1, thingAddr3)
	assert.False(t, isPublisher)
}
