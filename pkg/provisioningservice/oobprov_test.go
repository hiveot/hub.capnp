package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hiveot/hub/pkg/certservice/selfsigned"
	"github.com/hiveot/hub/pkg/provisioningservice/oobprovserver"
)

// Test creating and deleting the history database
// This requires a local unsecured MongoDB instance
func TestStartStop(t *testing.T) {
	caCert, caKey, err := selfsigned.CreateHubCA(1)
	assert.NoError(t, err)

	svc, err := oobprovserver.NewOobProvServer(caCert, caKey)
	assert.NoError(t, err)

	if assert.NotNil(t, svc) {
		assert.NoError(t, err)
	}
	assert.NoError(t, err)
}
