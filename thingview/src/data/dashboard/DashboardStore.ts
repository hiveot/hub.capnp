import { reactive, readonly } from '@vue/reactivity';
import {nanoid} from "nanoid";

// key to store dashboard in localstorage
const storageKey:string = "dashboards"

export const TileTypeCard = "card"
export const TileTypeImage = "image"
export const TileTypeLineChart = "linechart"

/**
 * Dashboard tile item to display
 */
export interface IDashboardTileItem {
  /**
   * Optional override of the property name
   */
  label?: string
  /**
   * ID of the thing whose property to display
   */
  thingID: string
  /**
   * ID of the thing property to display
   */
  propertyID: string,
}

/**
 * Dashboard tile configuration
 */
export class DashboardTileConfig extends Object {
  /**
* ID of the tile to display. Used to match the layout id.
*/
  id: string = nanoid(5);

  /**
   * Title of dashboard tile when displaying
   * Default is the first property name
   */
  title?: string

  /**
   * Type of widget to display in this tile
   */
  type: string = TileTypeCard  //

  /**
   * Collection of Thing items to display, in order of appearance
   */
  items: IDashboardTileItem[] = []
}

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
  tiles: { [id: string]: DashboardTileConfig } = {}

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
    console.info("AddDashboard %s", dashboard.name)
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
  AddTile(dashboard: DashboardDefinition, tile: DashboardTileConfig) {
    console.info("AddTile %s to dashboard %s", tile.title, dashboard.name)
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
   * Delete a dashboard
   * @param dashboard to remove.
   */
  DeleteDashboard(dashboard: DashboardDefinition) {
    console.info("RemoveDashboard %s", dashboard.name)

    // FIXME: concurrency ?
    let index = this.data.dashboards.findIndex((db) => db.id == dashboard.id);
    if (index < 0) {
      console.error(`RemoveDashboard: Dashboard ${dashboard.name} not found`)
      return
    }
    this.data.dashboards.splice(index, 1)
    this.Save()
  }

  /**
   * Delete a dashboard tile
   * @param dashboard that contains the tile
   * @param tile to remove
   */
  DeleteTile(dashboard: DashboardDefinition, tile: DashboardTileConfig) {
    console.info("DeleteTile '%s' from dashboard '%s'", tile.title, dashboard.name)

    let index = this.data.dashboards.findIndex((db) => db.id == dashboard.id);
    if (index < 0) {
      console.error(`RemoveTile: Dashboard ${dashboard.name} not found`)
      return
    }
    // The given dashboard is immutable. Use in-store reactive update. 
    let dashboardInStore = this.data.dashboards[index]
    delete dashboardInStore.tiles[tile.id]

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
  UpdateTile(dashboard: DashboardDefinition, tile: DashboardTileConfig) {
    console.log("UpdateTile '%s' in dashboard %s'", tile.title, dashboard.name)
    let index = this.data.dashboards.findIndex((db) => db.id == dashboard.id);
    if (index < 0) {
      let msg = `UpdateTile: Dashboard '${dashboard.name}' with id '${dashboard.id}' does not exist in the store.')`
      console.error(msg)
      throw(new Error(msg))
    }
    let dash = this.data.dashboards[index]
    let newTile = JSON.parse(JSON.stringify(tile))
    dash.tiles[newTile.id] = newTile
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