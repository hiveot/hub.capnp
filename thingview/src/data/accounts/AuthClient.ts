import jwtDecode, { JwtPayload } from 'jwt-decode'

// Default port of the Hub authentication service
const DefaultPort = 8881
// DefaultJWTLoginPath for obtaining access & refresh tokens
const DefaultJWTLoginPath = "/auth/login"
// DefaultJWTRefreshPath for refreshing tokens with the auth service
const DefaultJWTRefreshPath = "/auth/refresh"
// DefaultJWTConfigPath for storing client configuration on the auth service
const DefaultJWTConfigPath = "/auth/config"


export class ResponseError extends Error {

  constructor(message: string) {
    super(message)
  }
  public errorCode: number = 0
}

export class UnauthorizedError extends ResponseError {
  constructor(message: string) {
    super(message)
    this.errorCode = 401
  }
}

/**
 * Login request message format. Must match that of the auth service authenticator
 */
interface LoginRequestMessage {
  login: string        // username/email
  password: string     // password
  rememberMe: boolean  // persist the refresh token in a secure cookie
}

// Client for connecting to a Hub authentication service
export default class AuthClient {
  private address: string = ""
  private port: number = DefaultPort
  private accountID: string = ""
  private _accessToken: string | null = null
  // private _refreshToken: string = ""
  private caCert: string = "" // in PEM format
  private loginID: string = ""
  // 
  private sessionKey;

  // Create the authentication client for the given hub
  // @param accountID for callbacks and session storage
  // @param address of the auth service to connect
  // @param port of the auth service to connect to
  constructor(accountID: string, address: string, port: number | undefined) {
    this.accountID = accountID
    this.address = address
    if (!!port) {
      this.port = port
    }
    this.sessionKey = 'accessToken-' + accountID
    this._accessToken = sessionStorage.getItem(this.sessionKey)

    const options = {
      // key: fs.readFileSync("/srv/www/keys/my-site-key.pem"),
      // cert: fs.readFileSync("/srv/www/keys/chain.pem")
    };
  }

  /**
   * httpsPost issue https request using its bearer accessToken in authorization header
   */
  private async httpsPost(path: string, jsonPayload: string): Promise<Response> {

    let url = "https://" + this.address + ":" + this.port.toString() + path
    const response: Promise<Response> = fetch(url, {
      method: 'POST',
      body: jsonPayload,
      // credentials: "same-origin",
      headers: {
        'Content-Type': 'application/json',
        // 'Authorization': 'bearer '+this._accessToken,
      },
    })
    // if (!response.ok) {
    //     console.error("Error: httpsPost: ", response.statusText)
    // } else if (response.status >= 400) {
    //     console.error("HTTPS error: "+ response.status+'-'+response.statusText);
    // } else {
    //     onSuccess(response)
    // }
    return response
  }

  // Return the access token for the hub or null if it doesn't exist
  public get accessToken(): string | null {
    return this._accessToken
  }


  // AuthenticateWithLoginID connects to the authentication server and requests JWT access and refresh
  // tokens using loginID/password authentication.
  //
  // This uses JWT authentication using the POST /login path with a Json encoded
  // JwtAuthLogin message as body.
  //
  // The server returns a JwtAuthResponse message with an access token. If rememberMe is used then
  // the server sets a secure cookie with a refresh token for use by the Refresh method.
  // The access token is used as bearer token in the Authentication header for followup requests.
  //
  // @param loginID to login with, for example the user's email
  // @param password to login with. This password is not stored and only used to obtain the tokens.
  // @param rememberMe stores the resulting refresh token in a secure cookie for use with Refresh()
  // This returns a promise that completes on success with the access token, or fails if the
  // credentials are invalid or the auth service cannot be reached.
  public async AuthenticateWithLoginID(loginID: string, password: string, rememberMe: boolean): Promise<string> {
    this.loginID = loginID
    let url = "https://" + this.address + ":" + this.port.toString() + DefaultJWTLoginPath

    let loginMessage: LoginRequestMessage = {
      login: loginID,
      password: password,
          // FIXME: use the rememberMe parameter after this works
      rememberMe: rememberMe
    } // rememberMe }
    let payload = JSON.stringify(loginMessage)

    return this.httpsPost(DefaultJWTLoginPath, payload)
      .then(response => {
        if (response.status == 401) {
          console.error("AuthClient.AuthenticateWithLoginID: Authentication Error", response.statusText)
          throw (new UnauthorizedError("Authentication Error"))
        } else if (response.status >= 400) {
          console.error("AuthClient.AuthenticateWithLoginID: Authentication failed", response.status)
          throw (new ResponseError("Authentication failed: " + response.statusText))
        }
        // convert the result to json
        return response.json()
      })
      .then(jsonResponse => {
        console.log("AuthClient.AuthenticateWithLoginID: Authentication as %s successful", loginID)
        this._accessToken = jsonResponse.accessToken
        sessionStorage.setItem(this.sessionKey, jsonResponse.accessToken)
        return jsonResponse.accessToken
      })
  }

  // IsExpired test if the user is still considered authenticated based on access token expiration.
  // 
  // This returns the remaining validity in seconds or 0 if the token has expired.
  public Expiry(): number {
    if (this._accessToken) {
      let decoded = jwtDecode<JwtPayload>(this._accessToken)
      // console.log("AuthClient.IsExpired: decoded=", decoded)
      let t1: number = decoded.exp ? decoded.exp : 0
      let t2 = Date.now() / 1000
      let remaining = Math.floor(t1 - t2)
      return remaining > 0 ? remaining : 0
    }
    return 0
  }

  // Renew the access and refresh tokens
  // This returns a promise that completes on success with a new access token, or fails if no
  // valid refresh token is held in the secure cookie.
  //
  // If refresh fails then AuthenticateWithLoginID() must be called to renew the tokens. This
  // requires a loginID and password which the user needs to supply.
  public async Refresh() {

    return this.httpsPost(DefaultJWTRefreshPath, "")
      .then(response => {
        if (response.status >= 400) {
          console.error("AuthClient.Refresh: Authentication Error", response.statusText)
          throw (new UnauthorizedError("Authentication Error"))
        }
        // response contains access and refresh tokens
        return response.json()
      })
      .then(jsonResponse => {
        console.log("AuthClient.Refresh: Authentication token refresh for user: %s", this.loginID)
        this._accessToken = jsonResponse.accessToken
        sessionStorage.setItem(this.sessionKey, jsonResponse.accessToken)
        return (jsonResponse.accessToken)
      })
  }

}

