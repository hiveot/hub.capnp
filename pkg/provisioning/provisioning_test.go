package provisioning_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
	"github.com/hiveot/hub/pkg/provisioning/service/oobprovserver"
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
