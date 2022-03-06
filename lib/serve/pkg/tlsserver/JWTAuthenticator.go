package tlsserver

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/serve/pkg/hubnet"
	"net/http"
)

// JwtClaims this is temporary while figuring things out
type JwtClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// JWTAuthenticator verifies issued JWT access token using the provided public key.
// See JWTIssuer for test cases of the authenticator.
// The application must use .AuthenticateRequest() to authenticate the incoming request using the
// access token.
type JWTAuthenticator struct {
	// Service certificate whose public key is used for token verification
	publicKey *ecdsa.PublicKey
}

// AuthenticateRequest validates the access token
// The access token is provided in the request header using the Bearer schema:
//
//   Authorization: Bearer <token>
//
// Returns the authenticated user and true if there is a match, of false if authentication failed
func (jauth *JWTAuthenticator) AuthenticateRequest(resp http.ResponseWriter, req *http.Request) (userID string, match bool) {

	accessTokenString, err := hubnet.GetBearerToken(req)
	if err != nil {
		// this just means JWT is not used
		logrus.Debugf("JWTAuthenticator: No bearer token in request %s '%s' from %s", req.Method, req.RequestURI, req.RemoteAddr)
		return "", false
	}
	jwtToken, claims, err := jauth.DecodeToken(accessTokenString)
	_ = claims
	if err != nil {
		// token needs a refresh
		logrus.Infof("JWTAuthenticator: Invalid access token in request %s '%s' from %s: %s",
			req.Method, req.RequestURI, req.RemoteAddr, err)
		return "", false
	}
	// TODO: verify claims: iat, iss, aud

	// hoora its valid
	logrus.Debugf("JWTAuthenticator. Request by %s authenticated with valid JWT token", jwtToken.Header)
	return claims.Username, true
}

// DecodeToken and return its claims
//
// If the token is invalid then claims will be empty and an error is returned
// If the token is valid but has an incorrect signature, the token and claims will be returned with an error
func (jauth *JWTAuthenticator) DecodeToken(tokenString string) (
	jwtToken *jwt.Token, claims *JwtClaims, err error) {

	claims = &JwtClaims{}
	jwtToken, err = jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			return jauth.publicKey, nil
		})
	if err != nil || jwtToken == nil || !jwtToken.Valid {
		return jwtToken, claims, fmt.Errorf("invalid JWT token. Err=%s", err)
	}
	err = jwtToken.Claims.Valid()
	if err != nil {
		return jwtToken, claims, fmt.Errorf("invalid JWT claims: err=%s", err)
	}
	claims = jwtToken.Claims.(*JwtClaims)

	return jwtToken, claims, nil
}

// NewJWTAuthenticator creates a new JWT authenticator
// publicKey is the public key for verifying the private key signature
func NewJWTAuthenticator(publicKey *ecdsa.PublicKey) *JWTAuthenticator {
	//publicKeyDer, _ := x509.MarshalPKIXPublicKey(pubKey)

	ja := &JWTAuthenticator{publicKey: publicKey}
	return ja
}
