
import {AccountRecord} from "@/data/AccountStore";
import MqttClient from "@/data/MqttClient";
import DirectoryClient from "@/data/DirectoryClient";
import AuthClient from "@/data/AuthClient";


type AccountConnection = {
    name: string
    id: string
    authClient: AuthClient
    // mqttClient: MqttClient
    directoryClient: DirectoryClient
}

// Manage account connections to MQTT and Directory services
export default class ConnectionManager {
    // accounts that are watched
    // active mqtt broker clients by account ID
    private connections: Map<string, AccountConnection>
    private started: boolean

    // Create a new connection manager
    constructor() {
        this.connections = new Map<string, AccountConnection>()
        this.started = false
    }

    // Connect or reconnect
    async Connect(account: AccountRecord) {
        this.Disconnect(account.id);
        let authClient = new AuthClient(account.address, account.authPort);

        let mqttClient = new MqttClient(
            // this.handleMqttConnected.bind(this),
            // this.handleMqttDisconnected.bind(this),
            this.handleMqttMessage.bind(this)
        )
        let dirClient = new DirectoryClient()

        let c: AccountConnection = {
            name:account.name,
            id:account.id,
            authClient: authClient,
            // mqttClient: mqttClient,
            directoryClient: dirClient,
        }

        console.log("ConnectionManager.Connect: Connecting to", account.address, "as", account.loginName)
        this.connections.set(c.id, c)
        // after login, connect to mqtt and directory client using the access token
        let password = "user1" // FIXME - for testing
        let p = authClient.ConnectWithLoginID(account.loginName, password)
            .then((accessToken:string)=>{
                    console.log("ConnectionManager.Connect: Authentication successful. Connecting to mqtt and directory")
                    // FIXME support login to mqtt with access token
                    mqttClient.Connect(account, password)
                    dirClient.Connect(account.address, account.directoryPort, accessToken)
                })
            .catch((err:Error)=>{
                console.error("ConnectionManager.Connect: failed to connect: ", err)
                throw(err.message)
            })
        return p;
    }

    // Re-connect all enabled accounts
    ConnectAll(accounts: Array<AccountRecord>) {
        accounts.map((item: AccountRecord) => {
            if (item.enabled) {
                this.Connect(item)
            }
        })
    }

    // Nr of authenticated connections
    get connectionCount(): number {
        let count = 0
        this.connections.forEach((connection: AccountConnection) => {
            if (connection.authClient && connection.authClient.IsConnected()) {
                count++
            }
        })
        return count
    }

    // Handle an incoming MQTT message
    handleMqttMessage(topic: string, payload:Buffer, accountId: string, retain: boolean): void {
        console.log("handleMqttMessage. topic:",topic)
    }

    // Handle authentication login. This obtains the auth tokens for use with mqtt and directory service
    handleAuthLogin() {

    }

    // Handle refresh of the authentication tokens
    handleAuthRefresh() {

    }

    Disconnect(accountId:string) {
        let connection = this.connections.get(accountId)
        if (connection) {
            console.log("AccountManager.Disconnect: Disconnecting account:", connection.name)
            // connection.mqttClient.Disconnect();
            connection.directoryClient.Disconnect();
            this.connections.delete(connection.id)
        }
    }

    // Close all existing connections (modify the map as we go)
    DisconnectAll() {
        for(let key of Array.from( this.connections.keys()) ) {
           this.Disconnect(key)
        }
    }

}