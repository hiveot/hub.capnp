<script lang="ts" setup>

import {h, Component, reactive} from "vue";
import {DashboardDefinition, DashboardStore, DashboardTileConfig, IDashboardTileItem} from "@/data/dashboard/DashboardStore";
import {useQuasar, QCard, QCardSection, QBtn, QToolbar, QToolbarTitle} from "quasar";
import CardWidget from './CardWidget.vue'
import {matContentCopy, matContentPaste, matCopyAll, matDelete, matEdit, matMenu} from "@quasar/extras/material-icons";
import { ThingStore } from "@/data/td/ThingStore";
import TMenuButton, { IMenuItem } from "@/components/TMenuButton.vue";
import EditTileDialog from "./EditTileDialog.vue";

const $q = useQuasar()

/**
 * Dashboard Grid Tile
 * Wrapper around grid-item to handle positions
 */
const props = defineProps<{
  /**
   * Dashboard store used to edit tiles
   */
  dashStore: DashboardStore

/**
   * Things data storage
   */
  thingStore: ThingStore

  /**
   * Tile to edit
   */
  tile: DashboardTileConfig

  /** 
   * Dashboard the tile belongs to
   */
  dashboard: DashboardDefinition
}>()


interface IWidgetTypes { [key:string]:any};
const widgetTypes:IWidgetTypes = {
  "card": CardWidget,
  "image": CardWidget,
  // TileTypeLineChart: CardWidget,
}

// const getWidgetComponent = (config:DashboardTileConfig, ts:ThingStore):any => {
//   try{
//   let c = widgetTypes[config.type]
//   return h(
//     CardWidget, props{tile:props.tile, thingStore:props.thingStore}
//   )} catch(e){
//     console.error("getWidgetComponent Exception:",e)
//   }
// }

/**
 * Submit an updated tile to the store
 */
const handleSubmitTile = (newTile:DashboardTileConfig) => {
    props.dashStore.UpdateTile(props.dashboard, newTile)
    $q.notify({position: 'top',type: 'positive',
      message: 'Tile '+props.tile.title+' has been saved.'
    })
}

// Tile header dropdown menu 
const menuItems:IMenuItem[] = [
  {id: "edit",  label: "Edit Tile", icon: matEdit},
  {id: "copy",  label: "Copy Tile Items", icon: matContentCopy, disabled:true},
  {id: "paste", label: "Paste Tile Items", icon: matContentPaste, disabled:true},
  {separator: true},
  {id: "delete", label: "Delete Tile", icon: matDelete },
]

// Show the add tile dialog
const handleEditTile = (config:DashboardTileConfig) => {
  console.log("Opening edit tile...");
  $q.dialog({
    component: EditTileDialog,
    componentProps: {
      title: "Edit Tile",
      tile: props.tile,
    },
  }).onOk((newTile:DashboardTileConfig)=> {
    props.dashStore.UpdateTile(props.dashboard, newTile)
    $q.notify({position: 'top',type: 'positive',
      message: 'Tile '+props.tile.title+' has been saved.'
    })
  })
}

// Show the delete tile confirmation dialog
const handleDeleteTile = () => {
  $q.dialog({
    title: 'Confirm Delete',
    message: "This will delete Tile '"+props.tile.title+"'. Please confirm",
    cancel: true,
  }).onOk(()=> {
    props.dashStore.DeleteTile(props.dashboard, props.tile)
    // delete props.dashboard.tiles[props.config.id]

    console.info("Dashboard tile %s deleted", props.tile.title)
    $q.notify({position: 'top',type: 'positive',
      message: 'Tile '+props.tile.title+' has been deleted.'
    })

  })
}

const handleMenuAction = (menuItem:IMenuItem) => {
  switch (menuItem.id) {
    case 'edit': {handleEditTile(props.tile)}
    case 'copy': {; break}
    case 'paste': {; break}
    case 'delete': {handleDeleteTile(); break}
  }
}

</script>

<!--Display a widget based on dashboard tile configuration-->
<template>
  <QCard class="dashboard-tile-card">

    <span v-if="props.tile" 
        style="display:flex; flex-direction:column; width:100%; height:100%"
        >
       <!--  title with menu -->
      <QToolbar  class="tile-header-section" >
        <QToolbarTitle  class="toolbar-header">
           {{props.tile.title}}
        </QToolbarTitle>
        
        <!-- The menu is not a draggable area in the dashboard grid 
           'no-drag-area' is defined in the DashboardView
        -->
        <TMenuButton flat dense  class="toolbar-header no-drag-area"
            :icon="matMenu" 
            :items="menuItems"
            @on-menu-action="handleMenuAction"
            />
      </QToolbar>

      <!--       Slot for the widget content-->
      <!--       To keep vertical scrolling within the slot, use flex column with overflow-->
      <QCardSection class="tile-content-section">
        <CardWidget :tile="props.tile" :thingStore="props.thingStore"/>
      </QCardSection>

    </span>
    <span v-else>
      Oops. Widget configuration is not provided. Nothing to display.
    </span>

  </QCard>

</template>

<style scoped>

.dashboard-tile-card {
  display: flex;
  flex-direction: column;
  width:100%;
  height:100%;
  padding: 0;
  margin:0;
}

.tile-header-section {
  background-color:  rgb(224, 229, 230);
  min-height: 20px;
  padding: 0;
}
.toolbar-header {
  font-size: 1.1em;
  min-height: 20px;
  padding: 4px;
  /*height: 1rem;*/
}

.tile-content-section {
  /* height: 100%; */
  display: flex;
  flex-direction: column; 
  height:100%;
  width:100%;
  overflow: auto;
  padding: 0;
}

</style>