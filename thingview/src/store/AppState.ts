// AppState store for reactive access to non persistent application state
import { defineComponent, reactive, readonly } from "vue";


class AppStateData extends Object {
  editMode: Boolean = false;
  pages: Array<string> = ['page1', 'page2', 'page3'];
  selectedPage: String = "";
}

export class AppState {
  protected state: AppStateData;

  constructor() {
    this.state = reactive(new AppStateData())
  }

  // Return the reactive state
  // note, this should be readonly but that doesn't work on Array for some reason
  public getState(): AppStateData {
    // return readonly(this.state);
    return this.state;
  }

  // Change the edit mode on (true) or off (false)
  SetEditMode(on: boolean) {
    this.state.editMode = on;
  }

  // Change the selected page to display 
  SetSelectedPage(name: string) {
    this.state.selectedPage = name;
  }

  SetPages(names: string[]) {
    this.state.pages = names;
  }
}

const appState = new AppState();

export default appState;