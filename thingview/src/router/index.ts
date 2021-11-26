import {createWebHistory, createRouter, RouteRecordRaw, RouteLocationRaw} from "vue-router";
import PageView from "@/views/page/PageView.vue"
import AccountsView from '@/views/accounts/AccountsView.vue'

import { hubAuth } from "@/data/HubAuth";

import {PagesPrefix, PagesRouteName, AccountsRouteName} from "@/data/AppState";


const routes: Array<RouteRecordRaw> = [
  {
    name: "home",
    path: "/",
    component: PageView,
    // beforeEnter: checkAuth,
  },
  {
    name: AccountsRouteName,
    path: "/accounts",
    component: AccountsView,
  },
  {
    name: PagesRouteName,
    path: PagesPrefix+"/:page",
    component: PageView,

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
