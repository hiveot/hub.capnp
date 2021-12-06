// AppState data for reactive access to non persistent application state
import { reactive, watch } from "vue";
import {matDashboard} from "@quasar/extras/material-icons";

// Router constants
export const PagesPrefix = "/page"
export const AccountsRouteName = "Accounts"
export const PagesRouteName = "Pages"

// load/save key
const storageKey:string = "appState"


export interface IPageRecord {
  label: string
  icon: string
  to: string
}


// The global application state
export class AppStateData extends Object {
  editMode: boolean = false;

// TODO move persistent pages configuration into its own data
  pages: Array<IPageRecord> = [
    {label:'Overview', to: PagesPrefix+'/overview', icon:matDashboard},
  ];
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

  public AddPage(record:IPageRecord) {
    this.state.pages.push(record)
  }


  // load state from local storage
  Load() {
    console.log("AppState.Loading state")
    let serializedState = localStorage.getItem(storageKey)
    if (serializedState != null) {
      let state = JSON.parse(serializedState)
      this.state.editMode = state.editMode
      this.state.pages.splice(0, this.state.pages.length)
      this.state.pages.push(...state.pages )
    }
  }

  public RemovePage(record:IPageRecord) {
    let index = this.state.pages.indexOf(record)
    if (index >= 0) {
      this.state.pages.splice(index, 1)
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