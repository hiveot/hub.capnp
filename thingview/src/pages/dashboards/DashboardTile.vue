<script lang="ts" setup>

import {h, Component} from "vue";
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
  config: DashboardTileConfig

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

const getWidgetComponent = (config:DashboardTileConfig, ts:ThingStore):any => {
  try{
  let c = widgetTypes[config.type]
  return h(
    CardWidget, 
    {config:config, ts:ts}
  )
  } catch(e){
    console.error("getWidgetComponent Exception:",e)
  }
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
      tile: props.config,
    },
  }).onOk((newTile:DashboardTileConfig)=> {
    props.dashStore.UpdateTile(props.dashboard, newTile)
    $q.notify({position: 'top',type: 'positive',
      message: 'Tile '+props.config.title+' has been saved.'
    })
  })
}

// Show the delete tile confirmation dialog
const handleDeleteTile = () => {
  $q.dialog({
    title: 'Confirm Delete',
    message: "This will delete Tile '"+props.config.title+"'. Please confirm",
    cancel: true,
  }).onOk(()=> {
    props.dashStore.DeleteTile(props.dashboard, props.config)
    // delete props.dashboard.tiles[props.config.id]

    console.info("Dashboard tile %s deleted", props.config.title)
    $q.notify({position: 'top',type: 'positive',
      message: 'Tile '+props.config.title+' has been deleted.'
    })

  })
}



const handleMenuAction = (menuItem:IMenuItem) => {
  switch (menuItem.id) {
    case 'edit': {handleEditTile(props.config); break}
    case 'copy': {; break}
    case 'paste': {; break}
    case 'delete': {handleDeleteTile(); break}
  }
}

</script>

<!--Display a widget based on dashboard tile configuration-->
<template>
  <QCard
    style="display:flex; flex-direction:column; width:100%; height:100%">

    <span v-if="props.config">
       <!--  title with menu -->
      <QToolbar  class="tile-header-bar" >
        <QToolbarTitle  class="tile-header">
           {{props.config.title}}
        </QToolbarTitle>
        <TMenuButton flat dense  class="tile-header no-drag-area"
            :icon="matMenu" 
            :items="menuItems"
            @on-menu-action="handleMenuAction"
            />
      </QToolbar>

      <!--       Slot for the widget content-->
      <!--       To keep vertical scrolling within the slot, use flex column with overflow-->
      <QCardSection 
        style="height: 100%; display:flex; flex-direction:column; overflow: auto"
      >
        <CardWidget :config="props.config" :ts="thingStore"/>
      </QCardSection>

    </span>
    <span v-else>
      Oops. Widget configuration is not provided. Nothing to display.
    </span>

  </QCard>

</template>

<style scoped>

.tile-header-bar {
  background-color: lightgray;
  min-height: 20px;
  padding: 0;
}
.tile-header {
  font-size: 1.1em;
  min-height: 20px;
  padding: 4px;
  /*height: 1rem;*/
}
</style>