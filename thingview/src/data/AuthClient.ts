import jwt from 'jwt-simple'
import {matRememberMe} from "@quasar/extras/material-icons";

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
    private _accessToken: string = ""
    private _refreshToken: string = ""
    private caCert: string = "" // in PEM format
    private loginID: string = ""

    // Create the authentication client for the given hub
    // @param accountID for callbacks
    // @param address of the auth service to connect
    // @param port of the auth service to connect to
    constructor(address: string, port: number|undefined) {
        this.address = address
        if (!!port) {
            this.port = port
        }

        const options = {
            // key: fs.readFileSync("/srv/www/keys/my-site-key.pem"),
            // cert: fs.readFileSync("/srv/www/keys/chain.pem")
        };
    }

    // issue https request
    private async httpsPost(path:string, jsonPayload:string):Promise<Response> {
        // let options = {
        //     hostname: this.address,
        //     port: this.port,
        //     path: path,
        //     method: 'POST',
        //     ca: this.caCert,
        //     body: jsonPayload,
        // }

        let url = "https://"+this.address+":"+this.port.toString()+path
        const response:Promise<Response> = fetch(url, {
            method: 'POST',
            body: jsonPayload,
            // credentials: "same-origin",
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'bearer '+this._accessToken,
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

    // Return the access token for the hub "" if it doesn't exist
    public get accessToken():string {
        return this._accessToken
    }

    // Return the refresh token for the hub "" if it doesn't exist
    public get refreshToken():string {
        return this._refreshToken
    }

    // AuthenticateWithLoginID connects to the authentication server and requests JWT access and refresh
    // tokens using loginID/password authentication.
    //
    // This uses JWT authentication using the POST /login path with a Json encoded
    // JwtAuthLogin message as body.
    //
    // The server returns a JwtAuthResponse message with an access/refresh token pair and a refresh URL.
    // The access token is used as bearer token in the Authentication header for followup requests.
    //
    // @param loginID to login with, for example the user's email
    // @param password to login with. This password is not stored and only used to obtain the tokens.
    // @param rememberMe stores the resulting refresh token in a secure cookie
    // This returns a promise that completes on success or fails if the credentials are invalid
    // or the auth service cannot be reached
    public async AuthenticateWithLoginID(loginID: string, password: string, rememberMe: boolean):Promise<string> {
        this.loginID = loginID
        let url = "https://"+this.address+":"+this.port.toString()+DefaultJWTLoginPath
        // let data = JSON.stringify({login:loginID, password:password, rememberMe:true})
        // const httpsAgent = new https.Agent(
        //     {
        //         rejectUnauthorized: false
        //     });
        // let options = {
        //     url: url,
        //     data: data,
        //     headers: {
        //         'Content-Type': 'application/json',
        //         // 'Authorization': 'bearer ' + this._accessToken,
        //     },
        // }
        // return axios.post(url,data,options)
        //     // .then(results => results.json())
        //     .then(jsonResponse=>{
        //         console.log("AuthClient.Connect: Authentication successful")
        //         let responseMessage = JSON.parse(jsonResponse.data)
        //         this._accessToken = responseMessage.accessToken
        //         this._refreshToken = responseMessage.refreshToken
        //         return responseMessage.accessToken
        //     })
        //
        let loginMessage:LoginRequestMessage = {
            login: loginID,
            password: password,
            // FIXME: use the rememberMe parameter after this works
            rememberMe: true } // rememberMe }
        let payload = JSON.stringify(loginMessage)
        return this.httpsPost(DefaultJWTLoginPath, payload)
            .then(response => {
                if (response.status == 401) {
                    console.error("AuthClient: Authentication Error", response.statusText)
                    throw( new UnauthorizedError("Authentication Error"))
                } else if (response.status >= 400) {
                    console.error("AuthClient: Authentication failed", response.status)
                    throw( new ResponseError("Authentication failed: "+response.statusText))
                }
                // convert the result to json
                return response.json()
            })
            .then(jsonResponse=>{
                console.log("AuthClient.Connect: Authentication as %s successful", loginID)
                this._accessToken = jsonResponse.accessToken
                this._refreshToken = jsonResponse.refreshToken
                return jsonResponse.accessToken
            })
    }

    // Return true if the access token is valid
    // This returns 0 if the token is not valid or has expired
    // It returns the number of seconds of validity remaining if valid, useful to
    // know if the token must be refreshed.
    public IsAuthenticated(): number {
        if (this._accessToken) {
            let decoded = jwt.decode(this._accessToken, "", true)
            console.log("IsAuthenticated: decoded=", decoded)
            return 10
        }
        return 0
    }

    // Renew the access and refresh token pair
    // This returns a promise that completes on success, or fails if no valid refresh token is held.
    // If no refresh token is provided then assume that 'rememberMe' was last used and a token
    //  is stored in a secured cookie.
    //
    // If refresh fails then AuthenticateWithLoginID() must be called to renew the tokens. This
    // requires a loginID and password which the user needs to supply.
    //
    // @param refreshToken optional refresh token to use if rememberMe is not enabled
    public async Refresh(refreshToken?: string) {
        // TODO: use refresh token if provided
        return this.httpsPost(DefaultJWTRefreshPath, "")
            .then(response => {
                if (response.status >= 400) {
                    console.error("AuthClient: Authentication Error", response.statusText)
                    throw(new UnauthorizedError("Authentication Error"))
                }
                // response contains access and refresh tokens
                return response.json()
            })
            .then(jsonResponse=>{
                console.log("AuthClient.Connect: Authentication token refresh for user: %s", this.loginID )
                this._accessToken = jsonResponse.accessToken
                this._refreshToken = jsonResponse.refreshToken
                return (jsonResponse.accessToken)
            })
    }

}

