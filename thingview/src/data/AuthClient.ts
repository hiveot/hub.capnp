
// Client for connecting to a Hub authentication service
export default class AuthClient {

    // Return the access token for the hub "" if it doesn't exist
    public get accessToken():string {
        return ""
    }

    // Return the refresh token for the hub "" if it doesn't exist
    public get refreshToken():string {
        return ""
    }

    // Authenticate with the auth service. This obtains access and refresh tokens that can be used
    // to connect to other hub services. If the refresh token is renewed before it expires, no call to
    // Authenticate is needed to obtain a new access token.
    // This service will automatically refresh if a valid token pair is obtained. Calls to
    // other services should always use the latest access token.
    //
    // @param loginId to login with, usually the email
    // @param password to login with. This should not be stored and needs to be provided by the user
    // This returns a promise that completes on success or fails if the credentials are invalid
    // or the auth service cannot be reached
    public Authenticate(loginId: string, password: string): Promise<void> {
        return new Promise<void>( (resolve, reject) => {

        })
    }

    // Connect to the auth server and refresh the access and refresh tokens
    // Use Authenticate if this fails.
    // @param address of the hub to connect to.
    // @param port with the port number of the authentication service. 0 to use the default port 8881
    // This returns a promise that completes on success or fails if the existing refresh token is not valid
    async Connect(address: string, port: number): Promise<{access:string, refresh:string}> {

    }

    // Remove the access token
    public Disconnect() {
    }

    // Return true if authentication has connected to the auth server and obtained an
    // access and refresh token
    public IsConnected(): boolean {
        return false
    }

    // Renew the access and refresh tokens
    // This returns a promise that completes on success, or fails if no valid refresh token is held.
    // If it fails, call Authenticate() to renew the tokens.
    public Refresh():Promise<void> {
        return new Promise<void>( (resolve, reject)=>{

        })
    }

}

