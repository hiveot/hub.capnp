// AppState data for reactive access to non persistent application state
import { reactive, watch } from "vue";
import {matDashboard} from "@quasar/extras/material-icons";



// localstorage load/save key
const  storageKey:string = "appState"


// The global application state
export class AppStateData extends Object {
  editMode: boolean = false;
}

// The non-persistent runtime application state is kept here
export class AppState {
  protected state: AppStateData;

  constructor() {
    this.state = reactive(new AppStateData())
    watch(this.state, ()=>{
      this.Save()
    })
  }

  // load state from local storage
  Load() {
    console.log("AppState.Loading state")
    let serializedState = localStorage.getItem(storageKey)
    if (serializedState != null) {
      let state = JSON.parse(serializedState)
      this.state.editMode = state.editMode
    }
  }

  // Return the reactive state
  // note, this should be readonly but that doesn't work on Array for some reason
  public State(): AppStateData {
    // return readonly(this.state);
    return this.state;
  }

  // Change the edit mode on (true) or off (false)
  SetEditMode(on: boolean) {
    this.state.editMode = on;
  }

  // save to local storage
  Save() {
    console.log("AppState.Saving state")
    let serializedStore = JSON.stringify(this.state)
    localStorage.setItem(storageKey, serializedStore)
  }
}

const appState = new AppState();

export default appState;