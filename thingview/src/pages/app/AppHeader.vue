
<script lang="ts" setup>
import {reactive} from "vue";

import {useQuasar, QToggle} from 'quasar';
const $q = useQuasar()
import {matDashboard} from "@quasar/extras/material-icons";

import { MenuAbout, MenuEditMode, MenuAddDashboard, MenuDeleteDashboard, MenuEditDashboard} from './MenuConstants';
import AppMenu from './AppMenu.vue';
import AboutDialog from "./AppAboutDialog.vue";
import AddDashboardDialog from "@/pages/dashboards/AddDashboardDialog.vue";
import AppPagesBar from "./AppPagesBar.vue";
import TConnectionStatus from "@/components/TConnectionStatus.vue"
import {IMenuItem} from "@/components/TMenuButton.vue";

import appState, {AccountsRouteName, AppState, DashboardPrefix, IDashboardRecord} from '@/data/AppState'
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

// Show the add dashboard dialog
const handleAddDashboard = () => {
  console.log("Opening add dashboard...");
  $q.dialog({
    component: AddDashboardDialog,
    componentProps: {
      title: "Add Dashboard"
    },
    // cancel: true,
    // ok: true,
  }).onOk((newDashboard:IDashboardRecord)=> {
    props.appState.AddDashboard(newDashboard)
  })
}

// Show the edit dashboard dialog
const handleEditDashboard = (dashboard:IDashboardRecord) => {
  console.log("Opening edit dashboard for '"+dashboard.label+"'");
  $q.dialog({
    component: AddDashboardDialog,
    componentProps: {
      dashboard: dashboard,
      title: "Edit Dashboard"
    },
    // cancel: true,
    // ok: true,
  }).onOk((newDashboard:IDashboardRecord)=> {
    // TODO: user proper reactive store
    dashboard.label = newDashboard.label
  })
}


// Show the delete dashboard confirmation dialog
const handleDeleteDashboard = (dashboard: IDashboardRecord) => {
  $q.dialog({
    title: 'Confirm Delete',
    message: "This will delete dashboard '"+dashboard.label+"'. Please confirm",
    cancel: true,
  }).onOk(()=> {
    appState.RemoveDashboard(dashboard)
  })
}

const handleEditModeChange = (ev:any)=>{
  console.log("AppHeader: emit onEditModeChange")
  props.appState.SetEditMode(ev == true)
}

// Show the about dialog
const handleOpenAbout = () => {
  console.log("Opening about...");
  $q.dialog({
    component: AboutDialog,
  })
}


// handle Dialog and edit mode select
const handleMenuSelect = (menuItem:IMenuItem, dashboard?:IDashboardRecord) => {
  console.log("handleMenuSelect: ", menuItem);
  // These items require a dashboard
  if (dashboard) {
    if (menuItem.id == MenuDeleteDashboard) {
      handleDeleteDashboard(dashboard)
    }
    else if (menuItem.id == MenuEditDashboard) {
      handleEditDashboard(dashboard)
    } 
  }
// menu items that do not require a  dashboard
  if (menuItem.id == MenuAbout) {
    handleOpenAbout();
  } else if (menuItem.id == MenuEditMode) {
    handleEditModeChange(!currentState.editMode);
  } else if (menuItem.id == MenuAddDashboard) {
    handleAddDashboard();
  }
}




</script>

<template>
  <div class="header">

    <!-- <AboutDialog v-if="data.showAbout" 
      :visible="true"
      @onClosed='data.showAbout = false'/>

    <AddDashboardDialog 
      :visible="data.showAddPage"
      @onClosed='data.showAddPage=false'
      @onAdd="handleAddDashboard"/>
 -->
    <img alt="logo" src="@/assets/logo.svg" @click="handleOpenAbout"
         style="height: 40px;cursor:pointer; padding:5px;"
    />

    <AppPagesBar 
      :dashboards="currentState.dashboards" 
      :edit-mode="currentState.editMode"
      @onMenuSelect="handleMenuSelect"
      />

    <div style="flex-grow:1"/>

    <!-- Edit mode switching -->
    <QToggle 
      :model-value="currentState.editMode"
      @update:model-value="handleEditModeChange"
      label="Edit"
      inactive-color="gray"
    />

    <!-- Connection Status -->
<!--    <TButton  icon="mdi-link-off" flat tooltip="Connection Status & Configuration"/>-->
    <TConnectionStatus 
      :value="cm.connectionStatus"
      :to="{name: AccountsRouteName}"
    />

    <!-- Dropdown menu -->
    <AppMenu
      :dashboards="currentState.dashboards"
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

