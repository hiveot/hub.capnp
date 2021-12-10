// AppState data for reactive access to non persistent application state
import { reactive, watch } from "vue";
import {matDashboard} from "@quasar/extras/material-icons";

// Router constants shared between router and navigation components
export const DashboardPrefix = "/dashboard"
export const AccountsRouteName = "accounts"
export const DashboardRouteName = "dashboard"
export const ThingsRouteName = "things"

// localstorage load/save key
const storageKey:string = "appState"


export interface IDashboardRecord {
  label: string
  icon: string
  to: string
}


// The global application state
export class AppStateData extends Object {
  editMode: boolean = false;

// TODO move persistent pages configuration into its own data
  dashboards: Array<IDashboardRecord> = [
    { label: 'Overview', to: DashboardPrefix + '/overview', icon: matDashboard },
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

  public AddDashboard(record: IDashboardRecord) {
    this.state.dashboards.push(record)
  }


  // load state from local storage
  Load() {
    console.log("AppState.Loading state")
    let serializedState = localStorage.getItem(storageKey)
    if (serializedState != null) {
      let state = JSON.parse(serializedState)
      this.state.editMode = state.editMode
      this.state.dashboards.splice(0, this.state.dashboards.length)
      this.state.dashboards.push(...state.dashboards)
    }
    // ensure there is at least 1 dashboard
    if (this.state.dashboards.length < 1) {
      this.state.dashboards.push({
        label: 'Overview', to: DashboardPrefix + '/overview', icon: matDashboard
      })
    }
  }

  // Remove a dashboard
  public RemoveDashboard(record: IDashboardRecord) {
    let index = this.state.dashboards.indexOf(record)
    if (index >= 0) {
      this.state.dashboards.splice(index, 1)
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