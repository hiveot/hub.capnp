import { reactive, readonly } from '@vue/reactivity';
import { Loading } from 'quasar';
// import { ListenOptions } from 'net';
import Store from '../Store'

export class DashboardWidgetDefinition extends Object {
  title: string = "";
}

export class DashboardDefinition extends Object {
  name: string = "";
  id: string = "";
  widgets: DashboardWidgetDefinition = new DashboardWidgetDefinition();
}

// Storage of the dashboards
export class DashboardCollection extends Object {
  // id: string = "";
  dashboards = new Map<string, DashboardDefinition>();
}


// DirectoryStore implements the data of Thing Description records
export class DashboardStore {
  data: DashboardCollection

  constructor() {
    this.data = reactive(new DashboardCollection())
  }

  // Add a new dashboard
  // @param id of dashboard
  // @param name of dashboard
  Add(id: string, name: string) {
    let dd = new DashboardDefinition({
      id: id,
      name: name,
    })
    this.data.dashboards.set(dd.id, dd)
  }

  // Return the dashboard with the given id
  getDashboard(id: string | undefined): DashboardDefinition | undefined {
    if (!id) {
      return undefined
    }
    const dash = this.data.dashboards.get(id)
    if (!dash) {
      return undefined
    }
    return readonly(dash) as DashboardDefinition
  }

  // Return the list of dashboards
  GetDashboards(): DashboardDefinition[] {
    let newList = Array<DashboardDefinition>()
    this.data.dashboards.forEach(val => {
      newList.push(val)
    })
    return readonly(newList) as Array<DashboardDefinition>
  }

  // Update an existing dashboard
  Update(record: DashboardDefinition) {

  }

  // Load the dashboard definition from the store
  Load() {
  }

  // Save the new dashboard definition to store
  Save() {

  }
}

const dashboardStore = new DashboardStore();
export default dashboardStore;