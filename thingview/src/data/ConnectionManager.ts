
import {AccountRecord} from "@/data/AccountStore";
import MqttClient from "@/data/MqttClient";
import DirectoryClient from "@/data/DirectoryClient";
import AuthClient from "@/data/AuthClient";


type AccountConnection = {
    name: string
    id: string
    authClient: AuthClient
    mqttClient: MqttClient
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
    async Connect(account: AccountRecord): Promise<AccountConnection> {
        this.Disconnect(account.id);
        let authClient = new AuthClient(
            this.handleAuthLogin.bind(this),
            this.handleAuthRefresh.bind(this)
        )
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
            mqttClient: mqttClient,
            directoryClient: dirClient,
        }

        console.log("ConnectionManager.Connect: Connecting to", account.address, "as", account.loginName)
        this.connections.set(c.id, c)
        try {
            let authResult = await authClient.Connect(account.address, account.authPort)
            console.log("ConnectionManager.Connect: Authentication successful")

            await mqttClient.Connect(account, authResult.access)
            await dirClient.Connect(account.address, account.directoryPort, authResult.access)
        } catch (err) {
            console.log("Authentication failed: ", err)
        }
        return c
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
            if (connection.authClient && connection.authClient.IsConnected) {
                count++
            }
        })
        return count
    }

    // Handle an incoming MQTT message
    handleMqttMessage(topic: string, payload:Buffer, accountId: string, retain: boolean): void {
        console.log("handleMqttMessage. topic:",topic)
    }

    Disconnect(accountId:string) {
        let connection = this.connections.get(accountId)
        if (connection) {
            console.log("AccountManager.Disconnect: Disconnecting account:", connection.name)
            connection.mqttClient.Disconnect();
            connection.directoryClient.Disconnect();
            connection.authClient.Disconnect();
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