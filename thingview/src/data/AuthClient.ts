// import tls from 'tls'
// import {json, text} from "stream/consumers";
// import axios from "axios";
// import fs from "fs";
// import https from "https";

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

    // issue https get request
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

    // ConnectWithLoginID creates a connection with the server using loginID/password authentication.
    // If a CA certificate is not available then insecure-skip-verify is used to allow
    // connection to an unverified server (leap of faith).
    // Authenticate with the auth service. This obtains access and refresh tokens that can be used
    // to connect to other hub services.
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
    // @param loginID to login with, usually the email
    // @param password to login with. This should not be stored and needs to be provided by the user
    // This returns a promise that completes on success or fails if the credentials are invalid
    // or the auth service cannot be reached
    public async ConnectWithLoginID(loginID: string, password: string):Promise<string> {
        this.loginID = loginID
        let url = "https://"+this.address+":"+this.port.toString()+DefaultJWTLoginPath
        let data = JSON.stringify({login:loginID, password:password})
        // const httpsAgent = new https.Agent(
        //     {
        //         rejectUnauthorized: false
        //     });
        let options = {
            url: url,
            data: data,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'bearer ' + this._accessToken,
            },
        }
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
        let payload = JSON.stringify({login:loginID, password:password})
        return this.httpsPost(DefaultJWTLoginPath, payload)
            .then(response => {
                if (response.status == 401) {
                    console.error("AuthClient: Authentication Error", response.statusText)
                    throw( new UnauthorizedError("Authentication Error"))
                } else if (response.status >= 400) {
                    console.error("AuthClient: Authentication failed", response.status)
                    throw( new ResponseError("Authentication failed: "+response.statusText))
                }
                return response.json()
            })
            .then(jsonResponse=>{
                console.log("AuthClient.Connect: Authentication successful")
                this._accessToken = jsonResponse.accessToken
                this._refreshToken = jsonResponse.refreshToken
                return jsonResponse.accessToken
            })
    }

    // Return true if a connection to obtain a token pair has succeeded, either through login or token refresh
    public IsConnected(): boolean {
        return this._accessToken != ""
    }

    // Renew the access and refresh tokens
    // This returns a promise that completes on success, or fails if no valid refresh token is held.
    // If it fails, call Authenticate() to renew the tokens.
    public async Refresh() {
        return this.httpsPost(DefaultJWTRefreshPath, "")
    }

}

