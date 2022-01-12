import {
  createWebHistory,
  createRouter,
  RouteRecordRaw,
  RouteLocationRaw, RouteParams,
} from "vue-router";

import { h } from 'vue'
// import AccountsView from '@/pages/accounts/AccountsView.vue'
import DialogRouterView from './DialogRouterView.vue'
import {ThingTD} from "@/data/td/ThingTD";
import dirStore from '@/data/td/ThingStore'
import { hubAuth } from "@/data/HubAuth";
import ds from "@/data/dashboard/DashboardStore";
import accountStore, {AccountRecord} from "@/data/accounts/AccountStore";

// Router constants shared between router and navigation components
// Should this move to router?index.ts
export const DashboardPrefix = "/dashboard"
export const AccountsRouteName = "accounts"
export const DashboardRouteName = "dashboard"
export const ThingsRouteName = "things"

// Get the account of the given ID or undefined if the ID is not found
const getAccount = (id:string): AccountRecord|undefined => {
  let account = accountStore.GetAccountById(id)
  if (!account) {
    console.log("Router getAccount id: ", id, 'is new.')
    return
  }
  console.log("Router getAccount id: ", id, 'found.')
  return account
}
// Get the thing of the given ID or an empty TD if the ID is not found
const getTD = (id:string): ThingTD => {
  let td = dirStore.GetThingTDById(id)
  if (!td) {
    console.log("Router getTD id: ", id, 'is new.')
    return new ThingTD()
  }
  console.log("Router getTD id: ", id, 'found!')
  return td
}

// Return the dashboard for the current route
// If no dashboard name is specified, return the first dashboard
// const dashboardPropsFn = (route:any):any => {
//   let dashboardName = route.params.dashboardName
//   console.debug('dashboardPropsFn. Getting dashboard with name ', dashboardName)
//   let dash = ds.GetDashboardByName(dashboardName)
//   if (!dash) {
//     // Dashboard not found.
//     console.warn("dashboardPropsFn: dashboard %s not found. Redirecting to first valid dashboard", dashboardName)
//   }
//   return {
//     dashboard: dash
//   }
// }

// Router paths and components
// Use dynamic components to reduce chunk size
const routes: Array<RouteRecordRaw> = [
  // List of accounts
  {
    // name: AccountsRouteName,
    path: "/accounts",
    // use dynamic loading to reduce load waits
    // component: () => import("@/pages/accounts/AccountsView.vue"),
    component: DialogRouterView,  // webstorm shows an error incorrectly
    children: [
      {
        // Display the list of saved accounts if no additional parameters are provided
        name: AccountsRouteName,
        path: '',
        component: () => import("@/pages/accounts/AccountsView.vue"),
      },
      {
        // Display the list of accounts as background and a dialog showing the account details
        name: 'accounts.dialog',
        path: ':accountID',
        components: {
          default: () => import("@/pages/accounts/AccountsView.vue"),
          // name 'dialog' matches the second router-view in EmptyRouterView
          dialog: () => import("@/pages/accounts/EditAccountDialog.vue"),
        },
        props: {
          dialog: (route: any) => ({
            returnTo: {name: AccountsRouteName},
            account: getAccount(route.params.accountID)
          }),
        }
      },
    ]
  },
  // list of things
  {
    // This 'things' route supports nested dialogs
    // DialogRouterView displays both the things as the dialog
    //   name: ThingsRouteName,
      path: "/things",
      component: DialogRouterView,  // webstorm shows an error incorrectly
      children: [
        {
          // Display the list of things if no additional parameters are provided
          name: ThingsRouteName,
          path: '',
          component: () => import("@/pages/things/ThingsView.vue"),
        },
        {
        // Display the list of things as background and a dialog showing the Thing details
        name: 'things.dialog',
        path: ':thingID',
        components: {
          default: () => import("@/pages/things/ThingsView.vue"),
          // name 'dialog' matches the second router-view in EmptyRouterView
          dialog: () => import("@/pages/things/ThingDetailsDialog.vue"),
        },
        props: {
          dialog: (route:any) => ({
            returnTo: {name:ThingsRouteName},
            td: getTD(route.params.thingID)
          }),
        }
      }
    ],
  },
  // selected dashboard
  {
    name: DashboardRouteName,
    path: DashboardPrefix + "/:dashboardName",
    component: () => import("@/pages/dashboards/DashboardView.vue"),
    props: true,
    // props: (route) => ( {dashboard: ds.GetDashboardByName(route.params.dashboardName as string)} )
    // beforeEnter: checkAuth,
  },
  {
    // vue router 4 no longer keeps it simple
    // path: '*', redirect: '/',
    path: '/:pathMatch(.*)*', redirect: '/accounts'
  }

];

// checkAuth redirects routes that require authentication to the login page when not logged in
function checkAuth(to: RouteLocationRaw, from: RouteLocationRaw, next: any) {
  if (!hubAuth.getState().isAuthenticated) {
    next("/login");
  } else {
    next();
  }
}

const router = createRouter({
  history: createWebHistory(),
  routes,
});


export default router;
