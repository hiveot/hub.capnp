package tlsserver_test

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/lib/client/pkg/certs"
	"github.com/wostzone/hub/lib/serve/pkg/tlsserver"
	"net/http"
	"net/http/httptest"
	"testing"
)

// JWT token creation and authentication test cases

func TestCreateJWTToken(t *testing.T) {
	user1 := "user1"
	derBytes := testCerts.ServerCert.Certificate[0]
	serverX509, _ := x509.ParseCertificate(derBytes)
	// issue the tokens
	issuer := tlsserver.NewJWTIssuer("issuerName", testCerts.ServerKey,
		func(loginID string, pass string) bool {
			return false
		})
	accessToken, refreshToken, err := issuer.CreateJWTTokens(user1)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// authenticate with the access token
	jauth := tlsserver.NewJWTAuthenticator(certs.PublicKeyFromCert(serverX509))

	decodedToken, claims, err := jauth.DecodeToken(accessToken)
	require.NoError(t, err)
	assert.NotEmpty(t, decodedToken)
	assert.NotNil(t, claims)
	assert.Equal(t, user1, claims.Username)

	decodedToken, claims, err = jauth.DecodeToken(refreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, decodedToken)
	assert.NotNil(t, claims)
	assert.Equal(t, user1, claims.Username)
}

func TestJWTBadSigningKey(t *testing.T) {
	assert.Panics(t, func() {
		tlsserver.NewJWTIssuer("issuerName", nil, nil)
	})
}

func TestJWTIncorrectSigningUser(t *testing.T) {
	// issue the tokens
	issuer := tlsserver.NewJWTIssuer("", testCerts.ServerKey,
		func(loginID string, pass string) bool {
			return false
		})
	// userID is required
	accessToken, refreshToken, err := issuer.CreateJWTTokens("")
	require.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestJWTIncorrectVerificationKey(t *testing.T) {
	user1 := "user1"
	someoneElseSecret := certs.CreateECDSAKeys()

	// issue the tokens
	issuer := tlsserver.NewJWTIssuer("issuerName", testCerts.ServerKey,
		func(loginID string, pass string) bool {
			return false
		})
	accessToken, _, err := issuer.CreateJWTTokens(user1)

	// verification should fail using someone else's key
	jauth := tlsserver.NewJWTAuthenticator(&someoneElseSecret.PublicKey)
	decodedToken, claims, err := jauth.DecodeToken(accessToken)
	assert.Error(t, err)
	assert.NotEmpty(t, decodedToken)
	assert.NotNil(t, claims)
	assert.Equal(t, user1, claims.Username)
}

func TestJWTBadToken(t *testing.T) {

	// issue the tokens
	issuer := tlsserver.NewJWTIssuer("issuerName", testCerts.ServerKey,
		func(loginID string, pass string) bool {
			return false
		})
	_, _, err := issuer.DecodeToken("notatoken")
	assert.Error(t, err)
}

func TestLogin(t *testing.T) {
	var didCheckCred = false

	// issue the tokens
	issuer := tlsserver.NewJWTIssuer("issuerName", testCerts.ServerKey,
		func(loginID string, pass string) bool {
			didCheckCred = true
			return true
		})
	//expTime := time.Now().Add(time.Second * 100)
	//accessToken, _, err := issuer.CreateJWTTokens(user1, expTime)

	//jauth := tlsserver.NewJWTAuthenticator(&testCerts.ServerKey.PublicKey)
	body, _ := json.Marshal(map[string]string{
		"username": "user1",
		"password": "pass",
	})
	req, err := http.NewRequest("PUT", "someurl", bytes.NewBuffer(body))
	assert.NoError(t, err)
	resp := httptest.NewRecorder()
	issuer.HandleJWTLogin(resp, req)

	assert.True(t, didCheckCred)
	// TODO: check if response has tokens
}

func TestBadLogin(t *testing.T) {

	// issue the tokens
	issuer := tlsserver.NewJWTIssuer("issuerName", testCerts.ServerKey,
		func(loginID string, pass string) bool {
			return false
		})

	body := http.NoBody
	req, err := http.NewRequest("GET", "someurl", body)
	assert.NoError(t, err)
	resp := httptest.NewRecorder()
	issuer.HandleJWTLogin(resp, req)
}

func TestRefresh(t *testing.T) {

	// issue the tokens
	issuer := tlsserver.NewJWTIssuer("issuerName", testCerts.ServerKey, nil)
	_, refreshToken, err := issuer.CreateJWTTokens("user1")
	assert.NoError(t, err)

	req, err := http.NewRequest("PUT", "someurl", http.NoBody)
	req.Header.Add("Authorization", "Bearer "+refreshToken)
	resp := httptest.NewRecorder()
	issuer.HandleJWTRefresh(resp, req)

	var result = string(resp.Body.Bytes())
	assert.NotEmpty(t, result)
	statusCode := resp.Result().StatusCode
	assert.Equal(t, 200, statusCode)
}

func TestRefreshNoToken(t *testing.T) {
	// issue the tokens
	issuer := tlsserver.NewJWTIssuer("issuerName", testCerts.ServerKey, nil)

	// empty bearer token
	req, _ := http.NewRequest("PUT", "someurl", http.NoBody)
	req.Header.Add("Authorization", "Bearer ")
	resp := httptest.NewRecorder()
	issuer.HandleJWTRefresh(resp, req)

	var result = string(resp.Body.Bytes())
	assert.Empty(t, result)
	statusCode := resp.Result().StatusCode
	assert.Equal(t, http.StatusUnauthorized, statusCode)
}

func TestRefreshInvalidToken(t *testing.T) {
	// issue the tokens
	issuer := tlsserver.NewJWTIssuer("issuerName", testCerts.ServerKey, nil)

	// bad bearer token
	req, _ := http.NewRequest("PUT", "someurl", http.NoBody)
	req.Header.Add("Authorization", "Bearer badtoken")
	resp := httptest.NewRecorder()
	issuer.HandleJWTRefresh(resp, req)

	var result = string(resp.Body.Bytes())
	assert.Empty(t, result)
	statusCode := resp.Result().StatusCode
	assert.Equal(t, http.StatusUnauthorized, statusCode)
}

func TestBadAccessToken(t *testing.T) {
	// issue the tokens
	//issuer := tlsserver.NewJWTIssuer("issuerName", testCerts.ServerKey, nil)
	jauth := tlsserver.NewJWTAuthenticator(&testCerts.ServerKey.PublicKey)

	// bad bearer token
	req, _ := http.NewRequest("GET", "someurl", http.NoBody)
	req.Header.Add("Authorization", "Bearer badtoken")
	resp := httptest.NewRecorder()
	userId, match := jauth.AuthenticateRequest(resp, req)

	assert.Empty(t, userId)
	assert.False(t, match)
}
