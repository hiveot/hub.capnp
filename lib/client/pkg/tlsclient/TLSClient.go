// Package tlsclient with a simple TLS client helper with certificate and username password authentication
package tlsclient

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
)

// Authentication methods for use with ConnectWithLoginID
// Use AuthMethodDefault unless there is a good reason not to
const (
	AuthMethodBasic = "basic" // basic auth for backwards compatibility when connecting to non WoST servers
	AuthMethodNone  = ""      // disable authentication, for testing
	AuthMethodJwt   = "jwt"   // JSON web token for use with WoST server (default)
)

// The default paths for user authentication and configuration
const (
	// DefaultJWTLoginPath for obtaining access & refresh tokens
	DefaultJWTLoginPath = "/auth/login"
	// DefaultJWTRefreshPath for refreshing tokens with the auth service
	DefaultJWTRefreshPath = "/auth/refresh"
	// DefaultJWTConfigPath for storing client configuration on the auth service
	DefaultJWTConfigPath = "/auth/config"
)

// JwtAuthLogin defines the login request message to sent when using JWT authentication
type JwtAuthLogin struct {
	LoginID  string `json:"login"`          // typically the email
	Password string `json:"password"`
}

// JwtAuthResponse defines the login or refresh response
type JwtAuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	RefreshURL   string `json:"refreshURL"`
}

// TLSClient is a simple TLS Client with authentication using certificates or JWT authentication with login/pw
type TLSClient struct {
	// host and port of the server to connect to
	hostPort        string
	caCert          *x509.Certificate
	caCertPool      *x509.CertPool
	httpClient      *http.Client
	timeout         time.Duration
	checkServerCert bool

	// client certificate mutual authentication
	clientCert *tls.Certificate

	// jwt authentication, default is jwt using DefaultLoginPath
	authMethod string
	userID     string
	secret     string
	// JwtTokens with access and refresh tokens. The access token is passed as
	// bearer token with each Invoke request. The refresh token is used to
	// refresh both tokens. These tokens can be shared with clients that connect
	// to other Hub services as a single-signon solution.
	JwtTokens *JwtAuthResponse
}

// Certificate returns the client auth certificate or nil if none is used
func (cl *TLSClient) Certificate() *tls.Certificate {
	return cl.clientCert
}

// Close the connection with the server
func (cl *TLSClient) Close() {
	logrus.Infof("TLSClient.Close: Closing client connection")

	if cl.httpClient != nil {
		cl.httpClient.CloseIdleConnections()
		cl.httpClient = nil
	}
}

// ConnectNoAuth creates a connection with the server without client authentication
// Only requests that do not require authentication will succeed
func (cl *TLSClient) ConnectNoAuth() {
	tlsConfig := &tls.Config{
		RootCAs:            cl.caCertPool,
		InsecureSkipVerify: !cl.checkServerCert,
	}

	tlsTransport := http.DefaultTransport
	tlsTransport.(*http.Transport).TLSClientConfig = tlsConfig

	cl.httpClient = &http.Client{
		Transport: tlsTransport,
		Timeout:   cl.timeout,
	}
}

// ConnectWithClientCert creates a connection with the server using a client certificate for mutual authentication.
// The provided certificate must be signed by the server's CA.
//  clientCert client tls certificate containing x509 cert and private key
// Returns nil if successful, or an error if connection failed
func (cl *TLSClient) ConnectWithClientCert(clientCert *tls.Certificate) (err error) {
	var clientCertList = []tls.Certificate{}

	if clientCert == nil {
		err = fmt.Errorf("TLSClient.ConnectWithClientCert, No client key/certificate provided.")
		logrus.Error(err)
		return err
	}

	// test if the given cert is valid for our CA
	if cl.caCert != nil {
		opts := x509.VerifyOptions{
			Roots:     cl.caCertPool,
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}
		x509Cert, err := x509.ParseCertificate(clientCert.Certificate[0])
		if err == nil {
			// FIXME: TestCertAuth: certificate specifies incompatible key usage
			// why? Is the certpool invalid? Yet the test succeeds
			_, err = x509Cert.Verify(opts)
		}
		if err != nil {
			logrus.Errorf("ConnectWithClientCert: certificate verfication failed: %s", err)
			return err
		}
	}
	cl.clientCert = clientCert
	clientCertList = append(clientCertList, *clientCert)

	tlsConfig := &tls.Config{
		RootCAs:            cl.caCertPool,
		Certificates:       clientCertList,
		InsecureSkipVerify: !cl.checkServerCert,
	}

	tlsTransport := http.DefaultTransport
	tlsTransport.(*http.Transport).TLSClientConfig = tlsConfig

	cl.httpClient = &http.Client{
		Transport: tlsTransport,
		Timeout:   cl.timeout,
	}
	return nil
}

// ConnectWithLoginID creates a connection with the server using loginID/password authentication.
// If a CA certificate is not available then insecure-skip-verify is used to allow
// connection to an unverified server (leap of faith).
//
// This uses JWT authentication using the POST /login path with a Json encoded
// JwtAuthLogin message as body.
//
// The server returns a JwtAuthResponse message with an access/refresh token pair and a refresh URL.
// The access token is used as bearer token in the Authentication header for followup requests.
//
// If the access token is expired, the client will perform a refresh request using the refresh URL,
// before invoking the request.
//
// If AuthMethodNone is used, no authentication attempt will be made and the function will always be successful
//
// The behavior can be modified:
//  1. Alternate login URL by providing the 'authLoginURL' parameter
//  2. Alternate authentication method by adding the AuthMethod as 4th parameter:
//     - AuthMethodJwt: default. This will invoke the URL to obtain an authentication token from the server for further requests.
//     - AuthMethodNone: the server doesn't require authentication
//     - AuthMethodBasic: each future request will include basic authentication with the given credentials.
//
//  loginID username or application ID to identify as.
//  secret to authenticate with.
//  authLoginURL optional full address of the authentication server login, "" to authenticate using the application server /login
//  authMethod optional authentication method to use. Default is AuthMethodJwt
// Returns nil if successful or authMethod is AuthMethodNone, or an error if setting up of authentication failed.
func (cl *TLSClient) ConnectWithLoginID(loginID string, secret string,
	authLoginURL ...string) (accessToken string, err error) {
	cl.userID = loginID
	cl.secret = secret
	loginURL := fmt.Sprintf("https://%s%s", cl.hostPort, DefaultJWTLoginPath)

	if len(authLoginURL) > 0 && authLoginURL[0] != "" {
		loginURL = authLoginURL[0]
	}
	// AuthMethodNone or AuthMethodBasic can be used instead of the default AuthMethodJWT
	authMethod := AuthMethodJwt
	if len(authLoginURL) > 1 {
		authMethod = authLoginURL[1]
	}

	tlsConfig := &tls.Config{
		RootCAs:            cl.caCertPool,
		InsecureSkipVerify: !cl.checkServerCert,
	}
	// tlsTransport := http.Transport{
	// 	TLSClientConfig: tlsConfig,
	// }
	tlsTransport := http.DefaultTransport
	tlsTransport.(*http.Transport).TLSClientConfig = tlsConfig

	// FIXME:
	// 1 does this work if the server is connected using an IP address?
	// 2. How are cookies stored between sessions?
	cjarOpts := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	cjar, err := cookiejar.New(cjarOpts)
	if err != nil {
		logrus.Errorf("NewTLSClient: error setting cookiejar. The use of bearer tokens might not work: %s", err)
	}

	cl.httpClient = &http.Client{
		Transport: tlsTransport,
		Timeout:   cl.timeout,
		Jar:       cjar,
	}
	// Authenticate with JWT requires a cookiejar to store the refresh token
	if authMethod == AuthMethodJwt {

		loginMessage := JwtAuthLogin{
			LoginID:  loginID,
			Password: secret,
		}
		// resp, err2 := cl.Post(cl.jwtLoginPath, authLogin)
		resp, err2 := cl.Invoke("POST", loginURL, loginMessage)
		if err2 != nil {
			err = fmt.Errorf("ConnectWithLoginID: JWT login to %s failed. %s", loginURL, err2)
			return "", err
		}
		err2 = json.Unmarshal(resp, &cl.JwtTokens)
		if err2 != nil {
			err = fmt.Errorf("ConnectWithLoginID: JWT login to %s has unexpected response message: %s", loginURL, err2)
			return "", err
		}
		accessToken = cl.JwtTokens.AccessToken
	}
	// the authmethod is only valid after receiving a token
	cl.authMethod = authMethod
	return accessToken, err
}

// Delete sends a delete message with json payload
//  path to invoke
//  msg message object to include. This will be marshalled to json
func (cl *TLSClient) Delete(path string, msg interface{}) ([]byte, error) {
	// careful, a double // in the path causes a 301 and changes POST to GET
	url := fmt.Sprintf("https://%s%s", cl.hostPort, path)
	return cl.Invoke("DELETE", url, msg)
}

// Get is a convenience function to send a request
//  path to invoke
func (cl *TLSClient) Get(path string) ([]byte, error) {
	url := fmt.Sprintf("https://%s%s", cl.hostPort, path)
	return cl.Invoke("GET", url, nil)
}

// Invoke a HTTPS method and read response
// If authentication is enabled then add the auth info to the headers
//
//  method: GET, PUT, POST, ...
//  url: full URL to invoke
//  msg message object to include. Non strings will be marshalled to json
func (cl *TLSClient) Invoke(method string, url string, msg interface{}) ([]byte, error) {
	var body io.Reader = http.NoBody
	var err error
	var req *http.Request
	contentType := "application/json"

	if cl == nil || cl.httpClient == nil {
		logrus.Errorf("Invoke: '%s'. Client is not started", url)
		return nil, errors.New("Invoke: client is not started")
	}
	logrus.Infof("TLSClient.Invoke: %s: %s", method, url)

	// careful, a double // in the path causes a 301 and changes post to get
	// url := fmt.Sprintf("https://%s%s", hostPort, path)
	if msg != nil {
		// only marshal to JSON if this isn't a string
		switch msgWithType := msg.(type) {
		case string:
			body = bytes.NewReader([]byte(msgWithType))
		case []byte:
			body = bytes.NewReader(msgWithType)
		default:
			bodyBytes, _ := json.Marshal(msg)
			body = bytes.NewReader(bodyBytes)
		}
	}
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// use basic auth as fallback. WoST prefers JWT
	if cl.authMethod == AuthMethodBasic {
		if cl.userID != "" && cl.secret != "" {
			req.SetBasicAuth(cl.userID, cl.secret)
		}
	} else if cl.authMethod == AuthMethodJwt {
		if cl.JwtTokens.AccessToken != "" {
			cl.RefreshJWTTokenIfExpired()
			req.Header.Add("Authorization", "bearer "+cl.JwtTokens.AccessToken)
		}
	}

	// set headers
	req.Header.Set("Content-Type", contentType)

	resp, err := cl.httpClient.Do(req)
	if err != nil {
		logrus.Errorf("TLSClient.Invoke: %s %s: %s", method, url, err)
		return nil, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		msg := fmt.Sprintf("%s: %s", resp.Status, respBody)
		if resp.Status == "" {
			msg = fmt.Sprintf("%d (%s): %s", resp.StatusCode, resp.Status, respBody)
		}
		err = errors.New(msg)
	}
	if err != nil {
		logrus.Errorf("TLSClient:Invoke: Error %s %s: %s", method, url, err)
		return nil, err
	}
	return respBody, err
}

// Post a message with json payload
//  path to invoke
//  msg message object to include. Non strings will be marshalled to json
func (cl *TLSClient) Post(path string, msg interface{}) ([]byte, error) {
	// careful, a double // in the path causes a 301 and changes POST to GET
	url := fmt.Sprintf("https://%s%s", cl.hostPort, path)
	return cl.Invoke("POST", url, msg)
}

// Put a message with json payload
//  path to invoke
//  msg message object to include. Non strings will be marshalled to json
func (cl *TLSClient) Put(path string, msg interface{}) ([]byte, error) {
	// careful, a double // in the path causes a 301 and changes POST to GET
	url := fmt.Sprintf("https://%s%s", cl.hostPort, path)
	return cl.Invoke("PUT", url, msg)
}

// Patch sends a patch message with json payload
//  path to invoke
//  msg message object to include. Non strings will be marshalled to json
func (cl *TLSClient) Patch(path string, msg interface{}) ([]byte, error) {
	// careful, a double // in the path causes a 301 and changes POST to GET
	url := fmt.Sprintf("https://%s%s", cl.hostPort, path)
	return cl.Invoke("PATCH", url, msg)
}

// RefreshJWTTokens refreshes the JWT access and bearer token
//  refreshURL to use. "" for using the application server and default refresh path
// This returns a struct with new access and refresh token
func (cl *TLSClient) RefreshJWTTokens(refreshURL string) (refreshTokens *JwtAuthResponse, err error) {
	if refreshURL == "" {
		refreshURL = cl.JwtTokens.RefreshURL
	}
	if refreshURL == "" {
		refreshURL = fmt.Sprintf("https://%s%s", cl.hostPort, DefaultJWTRefreshPath)
	}

	// refresh token exists in client cookie
	req, err := http.NewRequest("POST", refreshURL, http.NoBody)
	var resp *http.Response
	if err != nil {
		logrus.Warningf("RefreshJWTTokens: Error creating request for URL %s: %s", refreshURL, err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = cl.httpClient.Do(req)

	if err != nil {
		logrus.Warningf("RefreshJWTTokens: Error using URL %s: %s", refreshURL, err)
		return nil, err
	} else if resp.StatusCode >= 400 {
		logrus.Warningf("RefreshJWTTokens: refresh using URL %s failed with: %s", refreshURL, resp.Status)
		return nil, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Infof("RefreshJWTTokens: failed with error %s", err)
		return cl.JwtTokens, err
	}
	err = json.Unmarshal(respBody, &cl.JwtTokens)
	return cl.JwtTokens, err
}

// RefreshJWTTokenIfExpired checks if the JWT access token is expired. If so, then refresh it.
// If the refresh token does not exist or the access token is not a JWT token then return
// without further action.
func (cl *TLSClient) RefreshJWTTokenIfExpired() {
	if cl.JwtTokens.RefreshToken == "" {
		return
	}

	if cl.JwtTokens.AccessToken != "" {
		claims := jwt.MapClaims{}
		_, _, err := new(jwt.Parser).ParseUnverified(cl.JwtTokens.AccessToken, &claims)
		if err != nil {
			// if the access token is invalid then don't do anything
			logrus.Warningf("RefreshJWTTokenIfExpired: Parse error on access token string: %s", err)
			return
		}
		err = claims.Valid()
		if err == nil {
			// the access token is still valid
			return
		}
	}
	cl.RefreshJWTTokens("")
}

// NewTLSClient creates a new TLS Client instance.
// Use Start/Stop to run and close connections
//  hostPort is the server hostname or IP address and port to connect to
//  caCert with the x509 CA certificate, nil if not available
// returns TLS client for submitting requests
func NewTLSClient(hostPort string, caCert *x509.Certificate) *TLSClient {
	var checkServerCert bool
	caCertPool := x509.NewCertPool()

	// Use CA certificate for server authentication if it exists
	if caCert == nil {
		logrus.Infof("NewTLSClient: destination '%s'. No CA certificate. InsecureSkipVerify used", hostPort)
		checkServerCert = false
	} else {
		logrus.Infof("TLSClient.NewTLSClient: destination '%s'. CA certificate '%s'",
			hostPort, caCert.Subject.CommonName)
		caCertPool.AddCert(caCert)
		checkServerCert = true
	}

	cl := &TLSClient{
		hostPort:        hostPort,
		timeout:         time.Second * 10,
		caCertPool:      caCertPool,
		caCert:          caCert,
		checkServerCert: checkServerCert,
		authMethod:      AuthMethodNone,
	}

	return cl
}
