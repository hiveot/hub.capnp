package idprovclient_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/lib/client/pkg/idprovclient"
)

// const discoveryAddr = "msi.local"

func TestSignature(t *testing.T) {

	message := "hello world"
	secret := "bob's secret"

	signature, err := idprovclient.Sign(message, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	// verify
	err = idprovclient.Verify(message, secret, signature)
	assert.NoError(t, err)
}

func TestSignatureDifferentSecret(t *testing.T) {

	message := "hello world"
	secret1 := "bob's secret"
	secret2 := "bob's secret2"

	signature, err := idprovclient.Sign(message, secret1)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	// verify should fail
	err = idprovclient.Verify(message, secret2, signature)
	assert.Error(t, err)
}

func TestSignatureDifferentMessage(t *testing.T) {

	message1 := "hello world"
	message2 := "hello world2"
	secret := "bob's secret"

	signature, err := idprovclient.Sign(message1, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	// verify should fail
	err = idprovclient.Verify(message2, secret, signature)
	assert.Error(t, err)
}

func TestSignatureInvalidEncoding(t *testing.T) {

	message := "hello world"
	secret := "bob's secret"

	signature, err := idprovclient.Sign(message, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	// verify should fail
	signature2 := signature + "."
	err = idprovclient.Verify(message, secret, signature2)
	assert.Error(t, err)
}
