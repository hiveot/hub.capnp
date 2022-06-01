package dirserver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wostzone/hub/thingdir/pkg/dirclient"
)

// This uses TestMain from DirServer_test

func TestGetThingPropertyValues(t *testing.T) {
	const Thing1ID = "thing1"
	const Prop1Name = "prop1"
	const Prop2Name = "prop2"
	const Prop1Value = "value1"
	const Prop2Value = 25.0

	// client for querying the result
	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)

	// publish a TD and update its properties
	propMap := make(map[string]interface{})
	propMap[Prop1Name] = Prop1Value
	// This adds thing1..4
	AddTds(directoryServer)
	directoryServer.UpdatePropertyValues(Thing1ID, propMap)
	assert.NoError(t, err)
	propMap[Prop2Name] = Prop2Value
	directoryServer.UpdatePropertyValues(Thing1ID, propMap)

	// Query the property values
	values, err := dirClient.GetThingValues(Thing1ID)
	assert.Equal(t, Prop1Value, values[Prop1Name].Value)
	assert.Equal(t, Prop2Value, values[Prop2Name].Value)

	// Query the events
	dirClient.Close()
}

func TestGetMultipleThingPropertyValues(t *testing.T) {
	const Thing1ID = "thing1"
	const Thing2ID = "thing2"
	const Prop1Name = "prop1"
	const Prop2Name = "prop2"
	const Prop1Value = "value1"
	const Prop2Value = 25.0

	// client for querying the result
	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)

	// publish a TD and update its properties
	propMap1 := map[string]interface{}{
		Prop1Name: Prop1Value,
	}
	// This adds thing1..4
	AddTds(directoryServer)
	directoryServer.UpdatePropertyValues(Thing1ID, propMap1)
	propMap2 := map[string]interface{}{
		Prop2Name: Prop2Value,
	}
	directoryServer.UpdatePropertyValues(Thing2ID, propMap2)

	// Query the property values
	thingIDs := []string{Thing1ID, Thing2ID, "notathing"}
	propNames := []string{Prop1Name, Prop2Name}
	props, err := dirClient.GetThingsPropertyValues(thingIDs, propNames)
	assert.Equal(t, Prop1Value, props[Thing1ID][Prop1Name].Value)
	assert.Equal(t, Prop2Value, props[Thing2ID][Prop2Name].Value)

	// Query the events
	dirClient.Close()
}
