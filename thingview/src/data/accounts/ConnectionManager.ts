import {reactive} from "vue";

import { AccountRecord } from "@/data/accounts/AccountStore";
import MqttClient from "./MqttClient";
import DirectoryClient from "@/data/td/DirectoryClient";
import AuthClient from "./AuthClient";

// Account connection status
export interface IConnectionStatus {
  readonly accountID: string        // ID of the account
  connected: boolean        // authenticated and at least one service connected
  authenticated: boolean    // authentication was successful
  directory: boolean        // the directory is obtained
  messaging: boolean        // message bus connection is established
  statusMessage: string     // human description of connection status
}
export type AccountConnection = {
  readonly accountID: string
  name: string
  authClient: AuthClient
  mqttClient: MqttClient
  dirClient: DirectoryClient
  state: IConnectionStatus
}

// Manage account connections to MQTT and Directory services
export class ConnectionManager {
  // accounts that are watched
  // active mqtt broker clients by account ID
  private connections: Map<string, AccountConnection>
  private started: boolean
  // active connection state
  private status: IConnectionStatus


  /** Create a new connection manager
    * @param onConnectChanged callback notifying of authentication and connection changes
    */
  constructor() {

    this.connections = new Map<string, AccountConnection>()
    this.status = reactive(<IConnectionStatus>{
      accountID: "",
      connected: false,
      authenticated: false,
      directory: false,
      messaging: false,
      statusMessage: "Not connected"
    })
    this.started = false
  }

  // Active connection status
  get connectionStatus(): IConnectionStatus {
    return this.status
  }

  // Nr of authenticated connections
  get connectionCount(): number {
    let count = 0
    this.connections.forEach((connection: AccountConnection) => {
      if (connection.authClient && connection.authClient.Expiry() > 0) {
        count++
      }
    })
    return count
  }


  /**
   * Authenticate the account with the authentication service and obtain an access token
   * Obtain new authentication tokens for the account using the given password
   *
   * @param account is the account to authenticate
   * @password to authenticate with
   * This returns a promise
   */
  async Authenticate(account: AccountRecord, password: string) {
    let ac = this.GetAccountConnection(account)
    if (!ac.authClient) {
      ac.authClient = new AuthClient(account.id, account.address, account.authPort)
    }
    console.log("ConnectionManager.Authenticate: Authenticate as", account.loginName, "to", account.address,)

    // AuthClient holds the tokens
    return ac.authClient.AuthenticateWithLoginID(account.loginName, password, account.rememberMe)
      .then((accessToken: string) => {
        console.log("ConnectionManager.Authenticate: Authentication successful")
        // save the auth status of the connection and overall status for presentation
        ac.state.authenticated = true
        ac.state.statusMessage = "Authentication successful"
        this.status.authenticated = true
      })
      .catch((err: Error) => {
        ac.state.authenticated = false
        ac.state.statusMessage = "Failed to authenticate: " + err.message
        console.error("ConnectionManager.Authenticate: failed to authenticate: ", err)
        throw (err.message)
      });
  }

  /**
   * Refresh authentication tokens.
   * This returns a promise that returns a new access token on success that can be used to
   * access other services.
   *
   * This only works on accounts with 'RememberMe' set as it relies on a secure cookie.
   *
   * This simply invokes the refresh API of the auth service. If a secure cookie was set containing
   * a valid refresh token, then both access and refresh tokens are renewed. Otherwise the call
   * will fail.
   *
   * @param account is the account to connect to authentication, mqtt and directory services
   * This returns a promise 
   */
  async AuthenticationRefresh(account: AccountRecord) {
    let ac = this.GetAccountConnection(account)
    if (!ac.authClient) {
      ac.authClient = new AuthClient(account.id, account.address, account.authPort)
    }
    console.log("ConnectionManager.AuthenticationRefresh: Refresh authentication with", account.address)
    this.connections.set(ac.accountID, ac)

    return ac.authClient.Refresh()
      .then((accessToken: string) => {
        console.log("ConnectionManager.AuthenticationRefresh: Authentication successful. Connecting to mqtt and directory")
        ac.state.authenticated = true
        ac.state.statusMessage = "Authentication refresh successful"
        this.status.authenticated = true
        return accessToken
      })
      .catch((err: Error) => {
        ac.state.authenticated = false
        ac.state.statusMessage = "Failed to refresh authentication token: " + err.message
        console.error("ConnectionManager.AuthenticationRefresh: failed to re-authenticate: ", err)
        throw (err.message)
      });
  }


    // Connect to Hub services using the access token obtained in authentication.
    // Authenticate or AuthenticationRefresh must be called first to obtain a valid token pair.
    //
    // This:
    //  1. Determine a valid access token
    //     a. in memory
    //     b. from session storage
    //     c. from refresh
    //  2. If a valid token exists then
    //     a. Load the directory from the directory service - publishers/things/??? limit the scope?
    //     b. Connect to the MQTT broker to receive real-time updates
    //  3. Invoke the call back with the result
    //
    //  @param account is the account to connect to mqtt and directory services
    //  @param onConnectChanged optional callback when the connection status changes
    // This returns a promise with access token 
    async Connect(account: AccountRecord,
      onConnectChanged: ((record: AccountRecord, status: IConnectionStatus) => void) | undefined = undefined) {
      this.Disconnect(account.id);

      // setup a new account connection or re-use existing one
      let ac = this.GetAccountConnection(account)

      console.log("ConnectionManager.Connect/1: address: ", account.address, "as", account.loginName)

      // if already connected then update status and continue
      if (ac.authClient.Expiry() > 1) {
        console.log("ConnectionManager.Connect/2a: Previous auth still valid for %s seconds ", ac.authClient.Expiry())
      } else {
        // Try refresh if no valid access token exists
        console.log("ConnectionManager.Connect/2b: Refreshing auth")
        // this throws when auth fails
        await this.AuthenticationRefresh(account)
      }

      ac.state.authenticated = true
      ac.state.statusMessage = "Authentication is valid"
      this.status.authenticated = true
      console.log("ConnectionManager.Connect/3: Connecting to mqtt and directory")

      if (ac.authClient.accessToken) {
        // Mqtt login accepts a valid access token
        if (ac.mqttClient && ac.authClient) {
          ac.mqttClient.Connect(account.loginName, ac.authClient.accessToken)
        }
        // if a directory client exists, get the directory
        if (ac.dirClient) {
          ac.dirClient.Connect(ac.authClient.accessToken)
        }
      }
      if (onConnectChanged) {
        onConnectChanged(account, ac.state)
      }
      return ac.authClient.accessToken

    }


    // Disable the connection
    Disconnect(accountId:string) {
        let connection = this.connections.get(accountId)
        if (connection) {
            console.log("AccountManager.Disconnect: Disconnecting account:", connection.name)
            if (connection.dirClient) {
                connection.dirClient.Disconnect();
            }
            if (connection.mqttClient) {
                connection.mqttClient.Disconnect();
            }
        }
    }


  /**
   * Get the connection of an account or create on if it doesn't exist
   * If an account connection isn't yet known, it will be created along with
   * clients for auth, mqtt and directory service.
   *
   * @param account to get the status of
   * returns the connection instance
   */
  protected GetAccountConnection(account: AccountRecord): AccountConnection {
    let connection = this.connections.get(account.id)
    if (!connection) {
      console.log("ConnectionManager.Connect/1: address: ", account.address, "as", account.loginName)

      connection = {
        name: account.name,
        accountID: account.id,

        authClient: new AuthClient(account.id, account.address, account.authPort),

        mqttClient: new MqttClient(
          account.id,
          account.address,
          account.mqttPort,
          this.handleMqttConnected.bind(this),
          this.handleMqttDisconnected.bind(this),
          this.handleMqttMessage.bind(this)),

        dirClient: new DirectoryClient(account.address, account.directoryPort),

        state: reactive(<IConnectionStatus>{
          statusMessage: "Not connected"
        })
      }
      this.connections.set(account.id, connection)
    }
    return connection
  }

  /**
   * Get the reactive connection status of an account.
   * The result is reactive and can be used directly in a UI
   * A status object will be created if it didn't yet exist.
   *
   * @param account whose status to get 
   * returns the connection status object
   */
  GetConnectionStatus(account: AccountRecord): IConnectionStatus {
    let connection = this.GetAccountConnection(account)
    return connection.state
  }

  /**
   * Handle an incoming MQTT message
   */
  handleMqttMessage(_accountID: string, topic: string, payload: Buffer, _retain: boolean): void {
    console.log("handleMqttMessage. topic:", topic, "Message size:", payload.length)

    // if a TD is received, update the Directory store

  }

  // track the MQTT account connection status
  handleMqttConnected(accountID: string) {
    let connection = this.connections.get(accountID)
    if (connection) {
      connection.state.messaging = true
      connection.state.connected = true
      connection.state.statusMessage = "Connected to message bus"
    }
    this.updateStatus()
  }
  // track the MQTT account connection status
  handleMqttDisconnected(accountID: string) {
    let connection = this.connections.get(accountID)
    if (connection) {
      connection.state.messaging = false
      connection.state.connected = false
      connection.state.statusMessage = "Disconnected from messaging"
    }
    this.updateStatus()
  }


    // update the aggregate connection status of all accounts
    updateStatus() {
      let messaging = false, connected = false, directory = false, authenticated = false;
        this.connections.forEach( c => {
            messaging = messaging || c.state.messaging
            connected = connected || c.state.connected
            directory = directory || c.state.directory
          authenticated = authenticated || c.state.authenticated
        })
        this.status.connected = connected
        this.status.messaging = messaging
        this.status.directory = directory
      this.status.authenticated = authenticated

      let newMessage = "Not connected"
      if (authenticated) {
        newMessage = "The user is authenticated"
        if (directory) {
          newMessage += ", the directory of Things is retrieved"
        }
        if (messaging) {
          newMessage += " and message bus connection is established"
        }
        if (!directory && !messaging) {
          newMessage = "Authenticated but not connected"
        }
      } else {
        newMessage = "Not authenticated"
      }
      this.status.statusMessage = newMessage

    }
}

// the global connection manager
const cm = new ConnectionManager()
export default cm