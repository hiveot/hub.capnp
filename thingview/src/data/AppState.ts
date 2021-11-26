// AppState data for reactive access to non persistent application state
import { reactive, readonly } from "vue";
import {mdiViewDashboard} from "@quasar/extras/mdi-v6";

// Router constants
export const PagesPrefix = "/page"
export const AccountsRouteName = "Accounts"
export const PagesRouteName = "Pages"


export interface IPageRecord {
  label: string
  icon: string
  to: string
}

// The global application state
// TODO move persistent pages configuration into its own data
export class AppStateData extends Object {
  editMode: boolean = false;
  pages: Array<IPageRecord> = [
    {label:'Overview', to: PagesPrefix+'/overview', icon:mdiViewDashboard},
  ];
}

// The non-persistent runtime application state is kept here
export class AppState {
  protected state: AppStateData;

  constructor() {
    this.state = reactive(new AppStateData())
  }

  public AddPage(record:IPageRecord) {
    this.state.pages.push(record)
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

}

const appState = new AppState();

export default appState;