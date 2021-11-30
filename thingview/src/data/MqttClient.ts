// import mqtt, {QoS, Packet} from 'mqtt'

// must use dist/mqtt.min see https://github.com/mqttjs/MQTT.js/issues/1269
import * as mqtt from 'mqtt/dist/mqtt.min'
import {QoS} from "mqtt";

const DefaultPort = 8885 // websocket protocol port. Use 8883 for mqtt protocol
const SYS_TOPIC = "$SYS/"
// MQTT $SYS topics
// export const SYS_CLIENTS_CONNECTED = "$SYS/broker/clients/connected"
// export const SYS_CLIENTS_DISCONNECTED = "$SYS/broker/clients/disconnected"
// export const SYS_CLIENTS_INACTIVE = "$SYS/broker/clients/inactive"
// export const SYS_CLIENTS_MAX = "$SYS/broker/clients/maximum"
// export const SYS_CLIENTS_TOTAL = "$SYS/broker/clients/total"
// export const SYS_BROKER_VERSION = "$SYS/broker/version"
// export const SYS_BROKER_UPTIME = "$SYS/broker/uptime"
// export const SYS_BROKER_SUBSCRIPTIONS = "$SYS/broker/subscriptions/count"
// export const SYS_MESSAGES_DROPPED = "$SYS/broker/publish/messages/dropped"
// export const SYS_MESSAGES_RECEIVED_TOTAL = "$SYS/broker/messages/received"
// export const SYS_MESSAGES_RECEIVED_PERMIN = "$SYS/broker/messages/received/1min"
// export const SYS_MESSAGES_STORED = "$SYS/broker/messages/stored"
// export const SYS_MESSAGES_RETAINED = "$SYS/broker/retained messages/count"
// export const SYS_BYTES_RECEIVED_PERMIN = "$SYS/broker/load/bytes/received/1min"
// export const SYS_BYTES_SENT_PERMIN = "$SYS/broker/load/bytes/sent/1min"


interface IMqttAccount {
    id: string
    address: string
    mqttPort?: number
    loginName: string
    // publishers: Array<{publisherID:string, subscribed:boolean}>
}

// Client for connecting to a Hub MQTT broker
export default class MqttClient {
    private accountInfo: IMqttAccount|null
    private connectedTimeStamp: number | null
    // private messageCount: number
    private mqttJS: mqtt.Client|null
    private onConnectedCallback?: (account: IMqttAccount, client: &MqttClient)=>void
    private onDisconnectedCallback?: (account: IMqttAccount)=>void
    private onMessageCallback: (topic: string, payload:Buffer, accountId: string, retain: boolean)=>void
    private msgCount:number
    private subscriptions: Array<any>
    private sysValues: Map<string,string> // map of broker $SYS topics and their values

    constructor(
        // onConnect: (account: IMqttAccount, client: &MqttClient)=>void,   // callback invoked when connected
        // onDisconnect: (account: IMqttAccount)=>void,
        onMessage: (topic: string, payload:Buffer, accountId: string, retained:boolean)=>void,
    ) {
        this.accountInfo = null
        this.msgCount = 0
        this.mqttJS = null
        // this.onConnectedCallback = onConnect
        // this.onDisconnectedCallback = onDisconnect
        this.onMessageCallback = onMessage

        // this.isConnected = false
        this.connectedTimeStamp = null
        // active subscription topics to resubscribe on connection restore
        this.subscriptions = []
        this.sysValues = new Map<string,string>()
    }

    /**
     * Connect to the MQTT broker
     */
    async Connect(accountInfo:IMqttAccount, accessToken: string) {
        if (this.isConnected) {
            this.Disconnect()
        }
        // Create a client instance
        // TODO: use template to populate server and port
        // let now = new Date()
        this.msgCount = 0
        this.accountInfo = accountInfo
        // let clientId = this.accountInfo.clientId ? this.accountInfo.clientId :
        //                "iotrain-dashboard-" + now.toISOString()
        // let port = this.accountInfo.port ? this.accountInfo.port : 1883

        //client = new Paho.MQTT.Client(mqtt_hostname, Number(mqtt_port), mqtt_client_id)
        // WebSockets use a different port. FIXME. let server handle connections
        // this.pahoClient = new Paho.Client(this.accountInfo.host, port, "", clientId)

        let mqttPort = accountInfo.mqttPort
        if (mqttPort == undefined || mqttPort == 0) {
            mqttPort = DefaultPort
        }
        let url = 'wss://' + this.accountInfo.address+":"+mqttPort.toString()
        let options:mqtt.IClientOptions = {
            reconnectPeriod: 3000,
            username: accountInfo.loginName,
            password: accessToken,
        }
        this.mqttJS = mqtt.connect(url, options);
        this.mqttJS.on('connect', this.handleConnected.bind(this))
        this.mqttJS.on('reconnect', this.handleReconnect.bind(this))

        this.mqttJS.on('disconnect', this.handleConnectionLost.bind(this))
        this.mqttJS.on('offline', this.handleConnectionLost.bind(this))
        this.mqttJS.on('error', this.handleConnectFailed.bind(this))

        this.mqttJS.on('message', this.handleMessage.bind(this))

        // this.messageCount = 0
        this.connectedTimeStamp = null

        // this.pahoClient.onConnectionLost = this.onConnectionLost.bind(this)
        // this.pahoClient.onMessageArrived = this.onMessage.bind(this)

        // connect the MQTT client
        // this.doConnect()
    }

    get ConnectedTimeStamp() {return this.connectedTimeStamp}

    /**
     * Disconnect if connected
     */
    Disconnect() {
        if (this.mqttJS != null && this.mqttJS.connected) {
            this.mqttJS.end( false, {}, () =>{
                if (this.onDisconnectedCallback) {
                    this.mqttJS = null
                    // Satisfy compiler check. A disconnect can only happen when accountInfo is set
                    if (this.accountInfo) {
                        this.onDisconnectedCallback(this.accountInfo)
                    }
                }
            })
        }
    }

    // // Establish connection to the MQTT broker
    // doConnect() {
    //   let connectOptions:ConnectionOptions = {
    //     onSuccess: this.handleConnected.bind(this),
    //     onFailure: this.handleConnectFailed.bind(this),
    //     // reconnect: true,
    //     timeout: 15,
    //     keepAliveInterval: 600,  // server disconnects when no activity for this amount of seconds
    //     cleanSession: true,
    //     invocationContext: this.accountInfo,
    //   }
    //   this.mqttClient.connect(connectOptions)
    // }

    get isConnected() {
        return this.mqttJS != null && this.mqttJS.connected
    }

    // The call to connect failed or timed out.
    // Invoke optional onDisconnectCallback and try again in 30 seconds
    handleConnectFailed(responseObject:Error) {
        console.warn("MqttClient.handleConnectFailed: Connection to MQTT broker failed: " + responseObject, "Retrying in 30 seconds...")

        this.subscriptions.length = 0
        if (this.onDisconnectedCallback) {
            // satisfy typescript
            if (this.accountInfo) {
                this.onDisconnectedCallback(this.accountInfo)
            }
        }
        // Wait 30 seconds before retrying
        // setTimeout(this.doConnect.bind(this), 30000)
    }

    // Connection was lost after initial connect succeeded.
    // Invoke optional callback. Connection will retry automatically
    handleConnectionLost() {
        this.subscriptions.length = 0
        console.warn("MqttClient.handleConnectionLost: Connection to MQTT broker lost")
        if (this.onDisconnectedCallback && this.accountInfo) {
            this.onDisconnectedCallback( this.accountInfo)
        }
        // paho client auto reconnects
        // setTimeout(this.doConnect.bind(this), 10000)
    }

    // Connection was established. Invoke optional callback and subscribe to messages
    handleConnected() {
        this.connectedTimeStamp = Date.now()
        console.log("MqttClient.handleConnected: Connection to MQTT broker established")
        // subscribe to basic $SYS info
        // if (this.accountInfo && this.accountInfo.subscribeToSys) {
        //     this.Subscribe([SYS_TOPIC+"#"])
        // }
        if (this.onConnectedCallback && this.accountInfo) {
            this.onConnectedCallback(this.accountInfo, this)
        }
    }
    // get sysTopicValues(): Map<string, string> {
    //     return this.sysValues
    // }

    // Message is received. Invoke optional callback
    handleMessage(topic:string, message:Buffer, packet:any){//Packet) {
        let t0 = performance.now();
        // this.messageCount++
        let retained:boolean = packet.retain // TODO is retain available?
        // let retained = false

        // console.log("MqttClient:onMessage: ", topic)
        // let payloadBytes = responseObject.payloadBytes
        if (topic.startsWith(SYS_TOPIC)) {
            this.sysValues.set(topic, message.toString())
        } else if (this.onMessageCallback) {
            let that = this
            // Don't block the next message, prevent dropping of messages?
            setTimeout(()=>{
                try {
                    if (that.accountInfo) {
                        this.onMessageCallback(topic, message, that.accountInfo.id, retained)
                    }
                } catch (err:any) {
                    console.error("MqttClient.handleMessage: Exception handling message on topic", topic, ". Stack trace:")
                    console.error(err.stack)   // this provides proper stack filename and line numbers
                }
            },0)
        }
        this.msgCount++
        let t1 = performance.now();
        console.log("MqttClient.handleMessage ("+topic+").", Math.round(t1 - t0) + " milliseconds. msgCount=",this.msgCount)
    }

    // reconnection starts
    handleReconnect() {
        console.log("MqttClient.handleReconnect. Trying to reconnect..." )
        // do nothing as mqtt client retains subscriptions
    }
    // publish a message on the mqtt bus
    Publish(topic:string, payload:string) {
        if (this.mqttJS) {
            this.mqttJS.publish(topic, payload, (err)=>{
                if (err) {
                    console.error("mqttjs-client:Publish Error", err)
                }
            })
        }
    }

    // subscribe to a topic. When following the home convention the data model is automatically updated.
    Subscribe(topics:Array<string>, qos:number=1) {
        // TODO: determine if subscriptions can be made while disconnected
        // let s = {topic:topic}
        // this.subscriptions.push(s)
        let subscribeOptions = {
            qos: qos as QoS,
        }
        // console.log("MqttClient.Subscribe: qos="+qos+" topics=",topics.toString())
        if (this.mqttJS) {
            this.mqttJS.subscribe(topics, subscribeOptions,
                (err, granted)=>{
                console.log("MqttClient.Subscribe to topic(s)",
                        granted ? "Success." : "Failed.", "Topic(s):", topics.toString())
            })
        }
    }

    // unsubscribe from a topic.
    Unsubscribe(topic:string) {
        // remove it from out tracked subscriptions list
        for (let index = 0; index < this.subscriptions.length; index++) {
            let s = this.subscriptions[index]
            try {
                if (s != null && s["topic"] == topic) {
                    this.subscriptions.splice(index,1)
                    break
                }
            } catch (err) {
                console.log("MqttClient.Unsubscribe: Error unsubscribe from topic '"+topic+"': ", err)
            }
        }
        if (this.mqttJS) {
            this.mqttJS.unsubscribe(topic)
        }
    }

}

