
<script lang="ts" setup>
import {reactive} from "vue";

import {QTabs, QRouteTab, QToggle} from 'quasar';
import {matDashboard} from "@quasar/extras/material-icons";

import AppMenu, { MenuAbout, MenuEditMode, MenuAddDashboard} from './AppMenu.vue';
import AboutDialog from "./AppAboutDialog.vue";
import AddPageDialog from "@/pages/dashboards/AddDashboardDialog.vue";
import TConnectionStatus from "@/components/TConnectionStatus.vue"
import {IMenuItem} from "@/components/MenuButton.vue";

import {AccountsRouteName, AppState, DashboardPrefix} from '@/data/AppState'
import cm, {IConnectionStatus} from "@/data/ConnectionManager";

interface IAppHeader {
  appState: AppState
  connectionStatus: IConnectionStatus
}
const props = defineProps<IAppHeader>()

// for convenience
const currentState = props.appState.State()

const emit = defineEmits([
    "onMenuSelect",
  ])

const data =reactive({
  showAbout: false,
  showAddPage: false,
})

const handleAddPage = (name:string) => {
  props.appState.AddDashboard({label:name, to:DashboardPrefix+'/'+name, icon:matDashboard});
  console.log("Added page: ",name)
}
const handleEditModeChange = (ev:any)=>{
  console.log("AppHeader: emit onEditModeChange")
  props.appState.SetEditMode(ev == true)
}
const handleOpenAbout = () => {
  console.log("Opening about...");
  data.showAbout = !data.showAbout;
}
const handleOpenAddPage = () => {
  console.log("Opening add page...");
  data.showAddPage = !data.showAddPage;
}
const handleAboutClosed = () => {
  // console.log("About closed...");
  data.showAbout = false;
}

// handle Dialog and edit mode select
const handleMenuSelect = (menuItem:IMenuItem) => {
  console.log("handleMenuSelect: ", menuItem);
  if (menuItem.id == MenuAbout) {
    handleOpenAbout();
  } else if (menuItem.id == MenuEditMode) {
    handleEditModeChange(!currentState.editMode);
  } else if (menuItem.id == MenuAddDashboard) {
    handleOpenAddPage();
  } else {
  }
}


</script>

<template>
  <div class="header">

    <AboutDialog :visible="data.showAbout"
                 @onClosed='handleAboutClosed'/>
    <AddPageDialog :visible="data.showAddPage"
                   @onClosed='data.showAddPage=false'
                   @onAdd="handleAddPage"/>

    <img alt="logo" src="@/assets/logo.svg" @click="handleOpenAbout"
         style="height: 40px;cursor:pointer; padding:5px;"
    />

    <!-- On larger screens show a tab bar for dashboard page -->
    <QTabs   inline-label indicator-color="green">
      <QRouteTab v-for="dashboard in currentState.dashboards"
             :label="dashboard.label"
             :icon="dashboard.icon"
             :to="(dashboard.to === undefined) ? '' : dashboard.to"
      />
    </QTabs>

    <div style="flex-grow:1"/>

    <!-- Edit mode switching -->
    <QToggle :model-value="currentState.editMode"
              @update:model-value="handleEditModeChange"
              label="Edit"
              inactive-color="gray"
    />

    <!-- Connection Status -->
<!--    <TButton  icon="mdi-link-off" flat tooltip="Connection Status & Configuration"/>-->
    <TConnectionStatus :value="cm.connectionStatus"
                      :to="{name: AccountsRouteName}"
    />

    <!-- Dropdown menu -->
    <AppMenu :dashboards="currentState.dashboards"
             :editMode="currentState.editMode"
             @onMenuSelect="handleMenuSelect"
    />

  </div>
</template>

<style>
/* Tab bar should have header background */
/*.p-tabmenu .p-tabmenu-nav {*/
/*  background: transparent !important;*/
/*}*/
/*.p-tabmenu .p-tabmenu-nav .p-tabmenuitem .p-menuitem-link {*/
/*  background:transparent !important;*/
/*}*/

.header {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: flex-start;
  /*font-size: large;*/
  gap: 10px;
  /*height: 46px;*/
  background-color: rgb(218, 229, 231);
}
</style>

