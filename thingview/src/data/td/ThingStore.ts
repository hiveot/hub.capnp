// The Thing store holds the discovered thing TD's
// This is updated from the directory (see DirectoryClient) and MQTT messages
import { reactive, readonly } from "vue";
import { cloneDeep as _cloneDeep, extend as _extend } from 'lodash-es'
import { ThingTD } from "./ThingTD";

// Directory data data
class TDCollection extends Object {
  index: Map<string, ThingTD> = new Map<string, ThingTD>();
  array: Array<ThingTD> = new Array<ThingTD>()
}


// DirectoryStore implements the data of Thing Description records
export class ThingStore {
  private data: TDCollection

  constructor() {
    // // remove default placeholder TD when complete
    // let testThing = new ThingTD()
    // testThing.id = "default"
    // testThing.description = "Hub Thing"
    // testThing["@type"] = "Computer"

    this.data = reactive(new TDCollection())
    // this.Add(testThing)
  }


  // Add or replace a new discovered thing to the collection
  Add(td: ThingTD): void {
    this.Update(td)
  }

  get all(): ThingTD[] {
    return this.data.array
  }

  // Get the ThingTD with the given id
  GetThingTDById(id: string): ThingTD | undefined {

    let td = this.data.index.get(id)
    if (!td) {
      return undefined
    }
    return readonly(td) as ThingTD
  }


  // Update/replace a new discovered ThingTD in the collection
  // This will do some cleanup on the TD to ensure the ID's are in place
  Update(td: ThingTD): void {
    let existing = this.data.index.get(td.id)
    let newTD = _cloneDeep(td)

    if (!existing) {
      // This is a new TD
      let newTD = _cloneDeep(td)
      this.data.array.push(newTD)
      this.data.index.set(newTD.id, newTD)
    } else {
      // update existing TD
      _extend(existing, td)
    }
  }
}

// Singleton instance
const dirStore = new ThingStore()

export default (dirStore)
