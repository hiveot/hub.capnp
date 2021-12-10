// directory holds the thing directory entries obtained from the directory service
// and updated from mqtt TD update messages
import { reactive, readonly } from "vue";
import { ThingTD } from "./ThingTD";

// Directory data data
class TDCollection extends Object {
  index: Map<string, ThingTD> = new Map<string, ThingTD>();
}


// DirectoryStore implements the data of Thing Description records
export class DirectoryStore {
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
    this.data.index.set(td.id, td)
  }

  // Return a list of all things
  GetAll(): ThingTD[] {
    let values = Array.from(this.data.index.values())
    return readonly(values) as ThingTD[]
  }

  // Get the account with the given id
  GetThingTDById(id: string): ThingTD | undefined {

    let td = this.data.index.get(id)
    if (!td) {
      return undefined
    }
    return readonly(td) as ThingTD
  }

  // update/replace a new discovered thing to the collection
  Update(td: ThingTD): void {
    this.data.index.set(td.id, td)
  }
}

// Singleton instance
const dirStore = new DirectoryStore()

export default (dirStore)
