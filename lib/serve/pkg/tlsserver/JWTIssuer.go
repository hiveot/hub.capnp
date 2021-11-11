package tlsserver

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/tlsclient"
	"github.com/wostzone/hub/lib/serve/pkg/hubnet"
	"net/http"
	"time"
)

const JwtRefreshCookieName = "authtoken"

// JWTIssuer creates JWT access and refresh tokens when a valid login/pw is provided.
//
// The login step verifies the given credentials using an external credentials handler and issues an access token
// and refresh token. The refresh token is also stored in a secure client cookie to save the hassle for secure storage
// by the client.
//
// Signing and verification use an asymmetric private/public key pair. The Hub uses the service
// private key for certificate generation and public key from the certificate for verification.
//
// In order to use JWT authentication, the client must do the following:
// 1. On first use, login through the login endpoint of the authentication service. This returns the
//    access and refresh tokens.
// 2. Place the access token in the authorization header of the service request. The service
//    will extract the token and verify it against the known public key.
// 3. Before the access token expires invoke the refresh endpoint and receive a new set of tokens.
// 4. Replace the access token in the authorization header again. Same as pt 2.
// 5. If access or refresh fails, prompt the user to log in again with credentials
// 6. If the user logs out, invoke the logout endpoint to remove the refresh token.
//
// With the use of login, refresh and logout comes the need for an endpoint for each purpose. Each
// endpoint must be attached to the handler via the router. For example:
//  > router.HandleFunc("/login", .HandleJWTLogin)  body=JwtAuthLogin{}
//  > router.HandleFunc("/logout", .HandleJWTLogout)  cookie=refresh token
//  > router.HandleFunc("/refresh", .HandleJWTRefresh)  cookie=refresh token
//

// JWTIssuer issues JWT tokens
type JWTIssuer struct {
	issuerName string

	// the credentials verification handler
	verifyUsernamePassword func(username, password string) bool

	signingKey      *ecdsa.PrivateKey
	verificationKey *ecdsa.PublicKey

	// These can be modified at will
	AccessTokenValidity  time.Duration
	RefreshTokenValidity time.Duration

	// optional callback when an expired token is used
	// expiredTokenAlert func(claims *JwtClaims)
}

// CreateJWTTokens creates a new access and refresh token pair containing the userID.
// The result is written to the response and a refresh token is set securely in a client cookie.
//  userID is the login ID of the user to whom the token is assigned and will be included in the claims
func (issuer *JWTIssuer) CreateJWTTokens(userID string) (accessToken string, refreshToken string, err error) {
	logrus.Infof("CreateJWTTokens for user '%s'", userID)
	accessExpTime := time.Now().Add(issuer.AccessTokenValidity)
	refreshExpTime := time.Now().Add(issuer.RefreshTokenValidity)
	if userID == "" {
		err = fmt.Errorf("CreateJWTTokens: Missing userID")
		return
	}

	// Create the JWT claims, which includes the username and expiry time
	accessClaims := &JwtClaims{
		Username: userID,
		StandardClaims: jwt.StandardClaims{
			Id:     userID,
			Issuer: issuer.issuerName,
			//Audience: "Hub services",
			Subject: "accessToken",
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: accessExpTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	// Declare the token with the algorithm used for signing, and the claims
	jwtAccessToken := jwt.NewWithClaims(jwt.SigningMethodES256, accessClaims)
	accessToken, err = jwtAccessToken.SignedString(issuer.signingKey)
	if err != nil {
		return
	}

	// same for refresh token
	refreshClaims := &JwtClaims{
		Username: userID,
		StandardClaims: jwt.StandardClaims{
			Id:      userID,
			Issuer:  issuer.issuerName,
			Subject: "refreshToken",
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: refreshExpTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	// Create the JWT string
	jwtRefreshToken := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims)
	refreshToken, err = jwtRefreshToken.SignedString(issuer.signingKey)
	return accessToken, refreshToken, err
}

// DecodeToken and return its claims
// Set error if token not valid
func (issuer *JWTIssuer) DecodeToken(tokenString string) (
	jwtToken *jwt.Token, claims *JwtClaims, err error) {

	claims = &JwtClaims{}
	jwtToken, err = jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			return issuer.verificationKey, nil
		})
	if err != nil || jwtToken == nil || !jwtToken.Valid {
		return nil, nil, fmt.Errorf("invalid JWT token. Err=%s", err)
	}
	err = jwtToken.Claims.Valid()
	if err != nil {
		return jwtToken, nil, fmt.Errorf("invalid JWT claims: err=%s", err)
	}
	claims = jwtToken.Claims.(*JwtClaims)

	return jwtToken, claims, nil
}

// HandleJWTLogin handles a JWT login POST request.
// Attach this method to the router with the login route. For example:
//  > router.HandleFunc("/login", HandleJWTLogin)
//
// The body contains provided userID and password
// This:
//  1. returns a JWT access and refresh token pair
//  2. sets a secure, httpOnly, sameSite refresh cookie with the name 'JwtRefreshCookieName'
func (issuer *JWTIssuer) HandleJWTLogin(resp http.ResponseWriter, req *http.Request) {
	logrus.Infof("HttpAuthenticator.HandleJWTLogin")

	loginCred := tlsclient.JwtAuthLogin{}
	err := json.NewDecoder(req.Body).Decode(&loginCred)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	// this is not an authentication provider. Use a callback for actual authentication
	match := issuer.verifyUsernamePassword(loginCred.LoginID, loginCred.Password)
	if !match {
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}

	refreshExpTime := time.Now().Add(issuer.RefreshTokenValidity)
	accessToken, refreshToken, err := issuer.CreateJWTTokens(loginCred.LoginID)
	// this can't really go wrong
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		logrus.Errorf("HttpAuthenticator.HandleJWTLogin: error %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	issuer.WriteJWTTokens(accessToken, refreshToken, refreshExpTime, resp)
}

// HandleJWTRefresh refreshes the access/refresh token pair
// Attach this method to the router with the refresh route. For example:
//  > router.HandleFunc("/refresh", HandleJWTRefresh)
//
// A valid refresh token must be provided in the client cookie or set in the authorization header
//
// This:
//  1. Return unauthorized if no valid refresh token was found
//  2. returns a JWT access and refresh token pair if the refresh token was valid
//  3. sets a secure, httpOnly, sameSite refresh cookie with the name 'JwtRefreshCookieName'
func (issuer *JWTIssuer) HandleJWTRefresh(resp http.ResponseWriter, req *http.Request) {
	logrus.Infof("HttpAuthenticator.HandleJWTRefresh")
	var refreshTokenString string

	// validate the provided refresh token
	cookie, err := req.Cookie(JwtRefreshCookieName)
	if err == nil {
		refreshTokenString = cookie.Value
	} else {
		refreshTokenString, err = hubnet.GetBearerToken(req)
	}
	// no refresh token found
	if err != nil || refreshTokenString == "" {
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}

	// is the token valid?
	_, claims, err := issuer.DecodeToken(refreshTokenString)
	if err != nil || claims.Id == "" {
		// refresh token is invalid. Authorization refused
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}

	refreshExpTime := time.Now().Add(issuer.RefreshTokenValidity)
	accessToken, refreshToken, err := issuer.CreateJWTTokens(claims.Id)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		logrus.Errorf("HttpAuthenticator.HandleJWTLogin: error %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	issuer.WriteJWTTokens(accessToken, refreshToken, refreshExpTime, resp)

}

// WriteJWTTokens writes the access and refresh tokens as response message and in a
// secure client cookie. The cookieExpTime should be set to the refresh token expiration time.
func (issuer *JWTIssuer) WriteJWTTokens(
	accessToken string, refreshToken string, cookieExpTime time.Time, resp http.ResponseWriter) error {

	// Set a client cookie for refresh "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	http.SetCookie(resp, &http.Cookie{
		Name:     JwtRefreshCookieName,
		Value:    refreshToken,
		Expires:  cookieExpTime,
		HttpOnly: true, // prevent XSS attack (client cant read value)
		Secure:   true, //
		// assume that the client service/website runs on the same server to use cookies
		SameSite: http.SameSiteStrictMode,
	})

	response := tlsclient.JwtAuthResponse{AccessToken: accessToken, RefreshToken: refreshToken}
	responseMsg, _ := json.Marshal(response)
	_, err := resp.Write(responseMsg)
	return err
}

// NewJWTIssuer create a new issuer of JWT authentication tokens using asymmetric keys.
//  issuerName of the service issuing the token
//  signingKey for generating tokens, or nil to generate a random 64 byte secret
//  verifyUsernamePassword is the handler that validates the login credentials
func NewJWTIssuer(
	issuerName string,
	signingKey *ecdsa.PrivateKey,
	verifyUsernamePassword func(loginID, secret string) bool,
) *JWTIssuer {
	if signingKey == nil {
		logrus.Panic("Missing signing key")
	}
	issuer := &JWTIssuer{
		verifyUsernamePassword: verifyUsernamePassword,
		signingKey:             signingKey,
		verificationKey:        &signingKey.PublicKey,
		issuerName:             issuerName,
		AccessTokenValidity:    15 * time.Minute,
		RefreshTokenValidity:   7 * 24 * time.Hour,
	}
	return issuer
}
