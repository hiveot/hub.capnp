// Constants for use by applications
import fs from 'fs'
import { ThingTD } from './ThingTD'
import dirStore, { ThingStore } from "./ThingStore"


const DefaultServiceName = "thingdir"
const DefaultPort = 8886

// paths with REST commands
const PathThings = "/things"           // list or query path
const PathThingID = "/things/{thingID}" // for methods get, post, patch, delete

// query parameters
const ParamOffset = "offset"
const ParamLimit = "limit"
const ParamQuery = "queryparams"

const DefaultLimit = 100
const MaxLimit = 1000



// Client for connecting to a Hub Directory service
export default class DirectoryClient {
  private hostPort: string = ""
  private _accessToken: string = ""
  private _refreshToken: string = ""
  private caCert: string = "" // in PEM format
  private store: ThingStore
  // private tlsClient: TLSSocket|null = null

  // Directory service client
  // @param address of directory service to connect to.
  // @param port the directory service is listening on. Use default port if not provided
  constructor(address: string, port: number | undefined) {
    if (!port) {
      port = DefaultPort
    }
    this.hostPort = address + ":" + port.toString()
    this.store = dirStore
    // this.tlsClient = null
  }

  /* Close the connection to the directory server
   */
  async Close() {
  }

  /* ConnectWithClientCert opens the connection to the directory server using a client certificate for authentication
   *  @param clientCertFile  client certificate to authenticate the client with the broker
   *  @param clientKeyFile   client key to authenticate the client with the broker
   */
  // public ConnectWithClientCert(tlsClientCert: tls.Certificate):void {
  // }

  /* ConnectWithLoginID open the connection to the directory server using a login ID and password for authentication
   */
  // ConnectWithLoginID(loginID: string, password: string): Error {
  //     return null
  // }

  // Connect open the connection to the directory server using an access token
  // and load the directory content.
  // @param accessToken
  async Connect(accessToken: string) {
    this._accessToken = accessToken
    this.LoadDirectory()
  }

  /* Delete a TD
   * @param id of the Thing Description document
   */
  // Delete(id: string) :void {
  // }

  // Disconnect from the directory
  Disconnect() {
    // this currently does nothing
  }

  // GetTD the ThingTD with the given ID
  //  id is the ThingID whose ThingTD to get
  // GetTD(id: string): ThingTD|undefined {
  //     return undefined
  // }

  /* ListTDs
   * Load the TDs from the directory starting at the offset. The result is limited to the nr of records 
   * provided with the limit parameter. The server can choose to apply its own limit, in which case 
   * the lowest value is used.
   * @param offset of the list to query from
   * @param limit result to nr of TDs. Use 0 for default.
   */
  async ListTDs(offset: number, limit: number): Promise<Array<ThingTD>> {
    // https://smallstep.com/hello-mtls/doc/client/axios
    // const httpsAgent = new https.Agent({
    //     // cert: fs.readFileSync('clientCert.pem'),
    //     // key: fs.readFileSync('clientKey.pem'),
    //     ca: fs.readFileSync('caCert.pem'),
    // });
    let url = "https://" + this.hostPort + PathThings
    // axios.get(url, {httpsAgent})
    // let options = {
    //     hostname: this.address,
    //     port: this.port,
    //     path: path,
    //     method: 'POST',
    //     ca: this.caCert,
    //     body: jsonPayload,
    // }

    const promise: Promise<Array<ThingTD>> = fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'bearer ' + this._accessToken,
      },
    }).then(resp => resp.json())
      .then(res => {
        if (!res) {
          let err = new Error("ListTDs: Missing content")
          throw (err)
        }
        console.log("Response content=", res)
        return res
      }).catch((reason: any) => {
        console.error("ListTD Error reading from %s: %s", url, reason)
        throw (reason)
      })
    return promise
  }


  // Load the directory of things and update the directory store
  async LoadDirectory() {
    let offset = 0
    let limit = 100

    // TODO: repeat until all things are collected
    return this.ListTDs(0, 0).then((things: ThingTD[]) => {
      console.log("Received directory update containing '%s' items", things.length)
      for (let td of things) {
        let parts = ThingTD.GetThingIDParts(td.id)
        td.zone = parts.zone
        td.deviceID = parts.deviceID
        td.publisher = parts.publisherID
        td.deviceType = parts.deviceType
        this.store.Update(td)
      }
    }).catch((reason: any) => {
      console.warn("Failed retrieving directory: ", reason)
      throw (reason)
    })
  }



  /* PatchTD changes a TD with the attributes of the given TD
   */
  async PatchTD(id: string, td: ThingTD) {
  }

  /* QueryTDs with the given JSONPATH expression
   * Returns a list of TDs matching the query, starting at the offset. The result is limited to the
   * nr of records provided with the limit parameter. The server can choose to apply its own limit,
   * in which case the lowest value is used.
   * @param jsonPath with the query expression
   * @param offset is the start index of the list to query from
   * @param limit limits the result to nr of TDs. Use 0 for default.
   */
  async QueryTDs(jsonpath: string, offset: number, limit: number): Promise<Array<ThingTD>> {
    return Array<ThingTD>()
  }

  /* UpdateTD fully replaces the TD with the given ID, eg create/update
   */
  async UpdateTD(id: string, td: ThingTD) {
  }
}

