// Definition of the Thing's TD, Thing Description document
// This consists of the TD itself with properties

import { WoTDataType, WoTProperties, WoTRequired } from "./Vocabulary"

// ThingIDParts describes the parts of how a Thing ID is constructed
// This is specific to WoST as WoT doesn't standardize it.
class ThingIDParts {
  // The zone a thing belongs to. Typically set by the publisher's configuration
  // The default is 'local'
  public zone?: string
  // The deviceID of the publishing device, usually a gateway or service that manages multiple Things
  public publisherID?: string
  // The deviceID which is unique to a publisher, or globally unique if no publisher is specified 
  public deviceID: string = ""
  // The type of device. Highly recommended for easy filtering
  public deviceType: string = ""
}


// Thing Description property describing a property of a Thing
export class TDAction extends Object {
  // name of action
  public title: string = ""

  // description of action 
  public description: string = ""

  // action input parameters
  public inputs = new Map<string, {
    WoTDataType?: string,
    WoTProperties?: Map<string, string>,
    WoTRequired?: boolean,
  }> ()
}

// Thing Description property describing output events of a Thing
export class TDEvent extends Object {
  // name of event
  public name: string = ""

  // event parameters
  public params = new Map<string, string>()
}

// Thing Description property describing a property of a Thing
export class TDProperty extends Object {
  public id: string = ""

  public atType: string = ""  // type of property wost:configuration, wost:output, wost:attr
  public title: string = "" // description of property
  public unit: string = ""  // unit of value 
  public value: string = ""  // current value of property
  public type: string = ""  // value type: string (default), number, bool
  public default: string = "" // default value of property
  public writable: boolean = false  // configuration is writable
}


// Thing description document
export class ThingTD extends Object {
  // Unique thing ID
  public id: string = "";

  // Human description for a thing
  public description: string = "";

  // Type of thing defined in the vocabulary
  public "@type": string = ""; // type of thing

  // Collection of properties of a thing
  public properties = new Map<string, TDProperty>();

  // Collection of actions of a thing
  public actions = new Map<string, TDAction>();

  // Collection of events (outputs) of a thing
  public events = new Map<string, TDEvent>();


  // Split the ID into its parts to determine zone, publisher, device ID and type
  public get ThingIDParts(): ThingIDParts {
    let parts = this.id.split(":")
    let tidp = new ThingIDParts()

    if ((parts.length < 2) || (parts[0].toLowerCase() != "urn")) {
      // not a conventional thing ID
      return tidp
    } else if (parts.length == 5) {
      // this is a full thingID  zone:publisher:deviceID:deviceType
      tidp.zone = parts[1]
      tidp.publisherID = parts[2]
      tidp.deviceID = parts[3]
      tidp.deviceType = parts[4]
    } else if (parts.length == 4) {
      // this is a partial thingID  zone:deviceID:deviceType
      tidp.zone = parts[1]
      tidp.deviceID = parts[2]
      tidp.deviceType = parts[3]
    } else if (parts.length == 3) {
      // this is a partial thingID  deviceID:deviceType
      tidp.deviceID = parts[1]
      tidp.deviceType = parts[2]
    } else if (parts.length == 2) {
      // the thingID is the deviceID
      tidp.deviceID = parts[1]
    }
    return tidp
  }
}
