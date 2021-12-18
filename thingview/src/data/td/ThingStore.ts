// The Thing store holds the discovered thing TD's
// This is updated from the directory (see DirectoryClient) and MQTT messages
import { reactive, readonly } from "vue";
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
    // remove when complete
    let testThing = new ThingTD()
    testThing.id = "default"
    testThing.description = "Hub Thing"
    testThing["@type"] = "Computer"

    this.data = reactive(new TDCollection())
    this.Add(testThing)
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
  Update(td: ThingTD): void {
    let existing = this.data.index.get(td.id)
    if (!existing) {
      // add new
      this.data.array.push(td)
      this.data.index.set(td.id, td)
    } else {
      // update existing, keep reactivity
      existing.description = td.description
      existing["@type"] = td["@type"]
      existing.properties = td.properties
    }
  }
}

// Singleton instance
const dirStore = new ThingStore()

export default (dirStore)
