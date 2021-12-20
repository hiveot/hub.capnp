import {
  createWebHistory,
  createRouter,
  RouteRecordRaw,
  RouteLocationRaw,
} from "vue-router";

import AccountsView from '@/pages/accounts/AccountsView.vue'
import DialogRouterView from './DialogRouterView.vue'
import {ThingTD} from "@/data/td/ThingTD";
import dirStore from '@/data/td/ThingStore'

import { hubAuth } from "@/data/HubAuth";
import {
  DashboardPrefix,
  ThingsRouteName,
  DashboardRouteName,
  AccountsRouteName
} from "@/data/AppState";


// Get the thing of the given ID or an empty TD if the ID is not found
const getTD = (id:string): ThingTD => {
  console.log("--- getTD id: ", id, '---')
  let td = dirStore.GetThingTDById(id)
  if (!td) {
    return new ThingTD()
  }
  return td
}

// Router paths and components
// Use dynamic components to reduce chunk size
const routes: Array<RouteRecordRaw> = [
  // List of accounts
  {
    name: AccountsRouteName,
    path: "/accounts",
    // use dynamic loading to reduce load waits
    component: () => import("@/pages/accounts/AccountsView.vue"),
  },
  {
    // This 'things' route supports nested dialogs
    // DialogRouterView displays both the things as the dialog
    //   name: ThingsRouteName,
      path: "/things",
      component: DialogRouterView, // ignore the error in webstorm
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
          dialog: (route:any) => ({to:{name:ThingsRouteName}, td: getTD(route.params.thingID)}),
        }
      }
    ],
  },
  {
    name: DashboardRouteName,
    path: DashboardPrefix + "/:page",
    component: () => import("@/pages/dashboards/DashboardView.vue"),

    // props, see: https://router.vuejs.org/guide/essentials/passing-props.html
    // boolean mode: when props is true use route.params as component props
    props: true,

    // object mode: when props is an object it is set as-is. For static props.
    // props: {
    //   page: route.params.page,
    // },


    // function mode: function that returns props, eg compute props
    // props: route => {page: route.params.page},

    // beforeEnter: checkAuth,
  },
  // {
  //   path: "/home",
  //   name: "Home",
  //   // for named views, set each props separately for each named view
  //   components: {
  //   //   default: Component1,
  //   //   sidebar: Sidebar,
  //   },
  //   props: {
  //   //   default: {}, // props of Component1
  //   //   sidebar: {}, // props of sidebar
  //   },
  // },
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
