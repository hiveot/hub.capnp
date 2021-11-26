// directory holds the thing directory entries obtained from the directory service
// and updated from mqtt TD update messages
import Store from './Store'

class DirProperty extends Object {
  name: string = "";
  value: string = "";
}

// Directory records to be stored 
class ThingDescriptionRecord extends Object {
  name: string = "";
  id: string = "";
  properties = new Map<string, DirProperty>();
}

// Directory data data
class ThingDescriptionCollection extends Object {
  index: Map<string, ThingDescriptionRecord> = new Map<string, ThingDescriptionRecord>();
}


// DirectoryStore implements the data of Thing Description records
class DirectoryStore extends Store<ThingDescriptionCollection> {

  protected data(): ThingDescriptionCollection {
    const sdata = new ThingDescriptionCollection()
    return sdata;
  }

}


export const dirStore: DirectoryStore = new DirectoryStore()