
<script lang="ts" setup>
import {reactive} from "vue";

import {useQuasar, QToggle} from 'quasar';
const $q = useQuasar()
import router from '@/router'
import {matDashboard} from "@quasar/extras/material-icons";

import {
  MenuAbout,
  MenuEditMode,
  MenuAddDashboard,
  MenuDeleteDashboard,
  MenuEditDashboard,
  MenuAddTile
} from './MenuConstants';

import {DashboardPrefix, AccountsRouteName} from '@/router/index'

import AppMenu from './AppMenu.vue';
import AboutDialog from "./AppAboutDialog.vue";
import EditDashboardDialog from "@/pages/dashboards/EditDashboardDialog.vue";
import EditTileDialog from "@/pages/dashboards/EditTileDialog.vue"
import AppPagesBar from "./AppPagesBar.vue";
import TConnectionStatus from "@/components/TConnectionStatus.vue"
import {IMenuItem} from "@/components/TMenuButton.vue";

import {ConnectionManager, IConnectionStatus} from "@/data/accounts/ConnectionManager";
import {DashboardDefinition, DashboardStore, DashboardTileConfig} from "@/data/dashboard/DashboardStore";
import {AppState} from '@/data/AppState'


interface IAppHeader {
  appState: AppState
  cm: ConnectionManager
  dashStore: DashboardStore
  connectionStatus: IConnectionStatus
}
const props = defineProps<IAppHeader>()

// for convenience
const currentState = props.appState.State()

// const emit = defineEmits([
//     "onMenuAction",
//   ])

const data =reactive({
  showAbout: false,
  showAddPage: false,
})

// Show the add dashboard dialog
const handleAddDashboard = () => {
  console.debug("Opening add dashboard...");
  $q.dialog({
    component: EditDashboardDialog,
    componentProps: {
      title: "Add Dashboard"
    },
    // cancel: true,
    // ok: true,
  }).onOk((newDashboard:DashboardDefinition)=> {
    props.dashStore.AddDashboard(newDashboard)
    // select the new dashboard
    router.push(DashboardPrefix+"/" +newDashboard.name)
  })
}
// Show the add tile dialog
const handleAddTile = (dashboard:DashboardDefinition) => {
  console.debug("Opening add tile...");
  $q.dialog({
    component: EditTileDialog,
    componentProps: {
      title: "Add Tile",
      tile: new DashboardTileConfig(),
    },
  }).onOk((newTile:DashboardTileConfig)=> {
    props.dashStore.AddTile(dashboard, newTile)
    $q.notify("A new Tile has been added to Dashboard "+dashboard.name)
  })
}
// Show the delete dashboard confirmation dialog
const handleDeleteDashboard = (dashboard: DashboardDefinition) => {
  $q.dialog({
    title: 'Confirm Delete',
    message: "This will delete dashboard '"+dashboard.name+"'. Please confirm",
    cancel: true,
  }).onOk(()=> {
    props.dashStore.DeleteDashboard(dashboard)
    let newDashName = DashboardPrefix+"/"+props.dashStore.dashboards[0]?.name
    console.log("handleDeleteDashboard: Changing route to ", newDashName)
    router.push(newDashName )
    $q.notify("Dashboard "+dashboard.name+ " has been deleted")
  })
}

// Show the edit dashboard dialog
const handleEditDashboard = (dashboard:DashboardDefinition) => {
  console.log("handleEditDashboard: Opening edit dashboard for '"+dashboard.name+"'");
  $q.dialog({
    component: EditDashboardDialog,
    componentProps: {
      dashboard: dashboard,
      title: "Edit Dashboard"
    },
    // cancel: true,
    // ok: true,
  }).onOk((newDashboard:DashboardDefinition)=> {
    props.dashStore.UpdateDashboard(newDashboard)
    let newDashName = DashboardPrefix+"/"+newDashboard.name
    console.log("Changing route to ", newDashName)
    router.push(newDashName )
  })
}

const handleEditModeChange = (ev:any)=>{
  console.log("AppHeader: emit onEditModeChange")
  props.appState.SetEditMode(ev == true)
}

// Show the edit tile dialog
const handleEditTile = (dashboard:DashboardDefinition, tile:DashboardTileConfig) => {
  console.log("handleEditTile: Opening edit tile...");
  $q.dialog({
    component: EditTileDialog,
    componentProps: {
      title: "Edit Tile",
      tile: tile,
    },
  }).onOk((newTile:DashboardTileConfig)=> {
    props.dashStore.UpdateTile(dashboard, newTile)
  })
}


// Show the about dialog
const handleOpenAbout = () => {
  console.log("handleOpenAbout: Opening about...");
  $q.dialog({
    component: AboutDialog,
  })
}


// handle Dialog and edit mode select
const handleMenuAction = (menuItem:IMenuItem, dashboard?:DashboardDefinition) => {
  console.info("AppHeader.handleMenuAction: ", menuItem);
  // These menu actions require a dashboard
  if (dashboard) {
    if (menuItem.id == MenuDeleteDashboard) {
      handleDeleteDashboard(dashboard)
    } else if (menuItem.id == MenuEditDashboard) {
      handleEditDashboard(dashboard)
    } else if (menuItem.id == MenuAddTile) {
      handleAddTile(dashboard)
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
      :dashboards="props.dashStore.dashboards"
      :edit-mode="currentState.editMode"
      @onMenuAction="handleMenuAction"
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
      :value="props.cm.connectionStatus"
      :to="{name: AccountsRouteName}"
    />

    <!-- Dropdown menu -->
    <AppMenu
      :dashboards="props.dashStore.dashboards"
      :editMode="currentState.editMode"
      @onMenuAction="handleMenuAction"
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

