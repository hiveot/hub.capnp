import { reactive, readonly } from '@vue/reactivity';
import {nanoid} from "nanoid";

// key to store dashboard in localstorage
const storageKey:string = "dashboards"

// Data definition of a dashboard tile
export class DashboardTile extends Object {
  id: string = nanoid(5);
  title: string = "";
}
//
// export interface IDashboardGridItem {
//   i: string,
//   x: number,
//   y: number,
//   w: number,
//   h: number
// }
// A responsive layout contain a list of items for each layout key
// export interface ILayout {
//   [key:string]: Array<{i: string,
//    x: number,
//    y: number,
//    w: number,
//    h: number}>
// }

export class DashboardDefinition extends Object {
  /**
   * Name of the dashboard for presentation and selection
   */
  name: string = "New Dashboard"

  /**
   * Unique ID of the dashboard
   */
  id: string = nanoid(5)

  /**
   * Tiles to show in the dashboard view
   */
  //  tiles: Map<string, DashboardTile> = new Map<string, DashboardTile>();
  tiles: any = {}

  /**
   * The layout of the tiles in the dashboard, as used and updated by vue-grid-layout.
   * Each responsive breakpoint "lg", "md", "sm", "xs", "xxs" has its own layout array as follows:
   *   {
   *      lg: [ {i:,x:,y:,w:,h:}, {...}, ...] }.
   *      md: [ {...}, ... ],
   *   },
   *  where 'i' is the unique key of the tile
   */
  // layouts: Map<string, Object[]> = new Map<string, Object[]>()
  layouts: {[key:string]:[]} = {}
  // layouts: Object = {}// {[key:string]:[]} = {}
}


// DirectoryStore implements the data of Thing Description records
export class DashboardStore {
  data: {dashboards:Array<DashboardDefinition>}

  /**
   * Create the Dashboard store instance
   */
  constructor() {
    this.data = reactive({
      dashboards: new Array<DashboardDefinition>()
    })
  }

  /**
   * Add a new dashboard
   * @param dashboard to add
   */
  AddDashboard(dashboard: DashboardDefinition) {
    console.log("AddDashboard %s", dashboard.name)
    let newDash = JSON.parse(JSON.stringify(dashboard))
    newDash.id = nanoid(8)
    this.data.dashboards.push(newDash)
    this.Save()
  }
  /**
   * Add a new dashboard tile
   * @param dashboard to add the tile to
   * @param tile to add
   */
  AddTile(dashboard: DashboardDefinition, tile: DashboardTile) {
    console.log("AddTile %s to dashboard %s", tile.title, dashboard.name)
    const dash = this.data.dashboards.find((item)=>item.id == dashboard.id)
    if (!dash) {
      console.error("AddTile: Dashboard with ID '"+dashboard.id+"' not found")
      return
    }
    let clone = JSON.parse(JSON.stringify(tile))
    dash.tiles[clone.id] = clone
    this.Save()
  }

  /**
   * Get the list of dashboards
   * This list is reactive and readonly
   */
  get dashboards(): readonly DashboardDefinition[] {
    return readonly(this.data.dashboards) as DashboardDefinition[]
  }

  /**
   * Return the dashboard with the given id
   * @param id of existing dashboard to get
   */
  GetDashboardByID(id: string | undefined): DashboardDefinition | undefined {
    if (!id) {
      return undefined
    }
    const dash = this.data.dashboards.find((item)=>item.id == id)
    if (!dash) {
      return undefined
    }
    return readonly(dash) as DashboardDefinition
  }

  /**
   * Return the dashboard with the given name
   * This returns undefined if no dashboard with the name exists
   * @param name of existing dashboard to get
   */
  GetDashboardByName(name: string): DashboardDefinition | undefined {
    if (!name) {
      return undefined
    }
    const dash = this.data.dashboards.find((item)=>item.name == name)
    if (!dash) {
      return undefined
    }
    return readonly(dash) as DashboardDefinition
  }

  /**
   * Load the dashboard definitions from the store
   */
  Load() {
    let serializedDashes = localStorage.getItem(storageKey)
    if (serializedDashes != null) {
      let dashes = JSON.parse(serializedDashes)
      this.data.dashboards.splice(0, this.data.dashboards.length)
      this.data.dashboards.push(...dashes)
    }
    // ensure there is at least 1 dashboard
    if (this.data.dashboards.length < 1) {
      let newDash = new DashboardDefinition()
      this.data.dashboards.push(newDash)
    }
    console.log(`DashboardStore.Load. Loaded ${this.data.dashboards.length} dashboard(s) from local storage`)
  }

  /**
   * Remove a dashboard
   * @param dashboard to remove.
   */
  RemoveDashboard(dashboard: DashboardDefinition) {
    console.log("RemoveDashboard %s", dashboard.name)

    let index = this.data.dashboards.findIndex((db)=>db.id == dashboard.id);
    if (index < 0) {
      console.error(`RemoveDashboard: Dashboard ${dashboard.name} not found`)
      return
    }
    this.data.dashboards.splice(index, 1)
    this.Save()
  }

  /**
   * Update an existing dashboard or add a new one
   * If the dashboard does not exist this will throw an error
   * @param dashboard to update
   * @constructor
   */
  UpdateDashboard(dashboard: DashboardDefinition) {
    console.log("UpdateDashboard %s", dashboard.name)
    let index = this.data.dashboards.findIndex((db)=>db.id == dashboard.id);
    if (index < 0) {
      let msg = `Dashboard '${dashboard.name}' does not exist in the store.')`
      console.error(msg)
      throw(new Error(msg))
    }
    let newDash = JSON.parse(JSON.stringify(dashboard))
    this.data.dashboards[index] = newDash
    this.Save()
  }

  /**
   * Update an existing tile
   * @param dashboard the tile belongs to
   * @param tile to update
   */
  UpdateTile(dashboard: DashboardDefinition, tile:DashboardTile) {
    console.log("UpdateTile '%s' in dashboard %s'", tile.title, dashboard.name)
    const dash = this.data.dashboards.find((item)=>item.id == dashboard.id)
    if (!dash) {
      let msg = `UpdateTile: Dashboard '${dashboard.name}' with id '${dashboard.id}' does not exist in the store.')`
      console.error(msg)
      throw(new Error(msg))
    }
    // const tileIndex = dash.tiles.findIndex( item=>item.id == tile.id)
    // if (tileIndex < 0) {
    //   let msg = `UpdateTile: Tile with ID '${tile.id}' not found for dashboard '${dashboard.name}'`
    //   console.error(msg)
    //   throw(new Error(msg))
    // }
    let newTile = JSON.parse(JSON.stringify(tile))
    dash.tiles.set(newTile.id, newTile)
    this.Save()
  }

  /**
   *  Save the new dashboard definition to store
   */
  Save() {
    console.log(`DashboardStore.Save. Saving ${this.data.dashboards.length} dashboards to local storage.`)
    let serializedStore = JSON.stringify(this.data.dashboards)
    localStorage.setItem(storageKey, serializedStore)

  }
}

const dashboardStore = new DashboardStore();
export default dashboardStore;