// store for dashboard definitions
import { ListenOptions } from 'net';
import Store from './Store'

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
  id: string = "";
  dashboards = new Map<string, DashboardDefinition>();
}


// DirectoryStore implements the store of Thing Description records
export class DashboardStore extends Store<DashboardCollection> {

  protected data(): DashboardCollection {
    const pages = new DashboardCollection()
    return pages;
  }

}


const dashboardStore = new DashboardStore();
export default dashboardStore;