<script lang="ts" setup>

// This is a POC implementation of the DashboardView based on vue3-grid-layout
// The problems with vue3-grid-layout are:
// 1. It is not possible to simply iterate the widgets and let the grid-layout manage
//    the placement and responsive layout. Instead, the layout needs to be iterated and
//    you're on own to make sure each widget has a layout defined. Add 5 responsive layouts
//    and the problem multiplies.
// 2. There is no method to load or import a responsive layout.
//    Its internal state has undocumented variables that need to be reset properly
// 3. There is no method to save or export the current responsive layout after changes.
//    The responsive layout has to be updated first to include the current layout.
// 4. There is no easy way to switch dashboards and update the layout
//    Ideally you bind a new responsive layout and done. That doesn't work though.

import {nextTick, onMounted, reactive, ref, watch} from "vue";
import {GridLayout, GridItem} from 'vue3-grid-layout'

import dashStore, {DashboardDefinition } from '@/data/dashboard/DashboardStore'
import thingStore from '@/data/td/ThingStore'
import appState from '@/data/AppState'
import DashboardTile from "./DashboardTile.vue";

/**
 * Dashboard view shows the dashboard with the given name
 */
const props = defineProps<{
    /**
     * Name of dashboard to display
     */
    dashboardName: string
}>()

const gridLayout = ref()

export interface ILayoutItem {
  i: string,
  x: number,
  y: number,
  w: number,
  h: number
}

const data = reactive({
  dashboard: dashStore.GetDashboardByName(props.dashboardName),
  currentLayout:  Array<ILayoutItem>(),
})

watch(()=>props.dashboardName, ()=>{
  console.log("DashboardView.watch(dashboardName)")
  data.dashboard = dashStore.GetDashboardByName(props.dashboardName)
  // updateDashboard(data.dashboard)
})

/**
 * FIXME: how to detect dashboard replacement in the store?
 */
watch( ()=>data.dashboard, 
       ()=>{
            console.info("DashboardView.watch(dashboard). Dashboard %s has updated!", data.dashboard?.name)
            updateDashboard(data.dashboard)
        },
        {deep:true} // watch options
     )

onMounted(()=>{
  console.log("DashboardView.onMounted")
  data.dashboard = dashStore.GetDashboardByName(props.dashboardName)
  updateDashboard(data.dashboard)
})

// After changing the dashboard this updates the layouts for all breakpoints
const updateDashboard = (dashboard:DashboardDefinition|undefined) => {
  // a new dashboard needs restoring the layouts of all breakpoints
  if (!dashboard || !dashboard.layouts) {
    console.error("updateDashboard: no dashboards or layouts. Ignored")
    return
  } else if (!gridLayout || !gridLayout.value) {
    console.warn('handleNewDashboard: ref gridLayout not initialized')
    return
  }
  console.info("DashboardView.handleNewDashboard. Start. Restoring all layouts from saved dashboard",
      dashboard.name, ". \nlayouts: ", dashboard.layouts)

  // reset the layout and trigger an update of the current layout
  // unfortunately can't set responsiveLayouts directly, so wait for next-tick
  // gridLayout.value.responsiveLayouts = {...dashboard.layouts}
  gridLayout.value.layouts = {...dashboard.layouts}

  // FIXME: this seems to be the trigger to re-initialize the layout. Is there an 'official' way?
  // nextTick so the grid property bindings are updated with the new responsiveLayout
  nextTick(()=>{
    // clearing lastBreakpoint prevents overwriting the new layout with the layout from the previous dashboard
    gridLayout.value.lastBreakpoint = null
    gridLayout.value.initResponsiveFeatures()
    gridLayout.value.responsiveGridLayout()
    console.log("DashboardView.handleNewDashboard. Done")
  })
}

/**
 * handleBreakpointChange stores the current layout for the current breakpoint
 * and reloads the new layout.
 * This already exists in the GridLayout code at line 418 but has a bug? that existing
 * layouts are not replaced.
 * @param newBreakpoint: name of new layout breakpoint: xxs, xs, sm, md, lg
 * @param newBPLayout: new layout that is to be used. This will be fixed first
 */
const handleBreakpointChange = (newBreakpoint:string, newBPLayout:any) => {
  console.log("DashboardView.handleBreakpointChange: newBreakpoint ", newBreakpoint, ". New layout:", newBPLayout)

  // debugger
  let lastBreakpoint:string = gridLayout.value.lastBreakpoint
  if (lastBreakpoint ) {
    // Workaround: GridLayout hasn't saved the previous layout
    gridLayout.value.layouts[lastBreakpoint] = data.currentLayout
  }
  // Make sure the new layout contain all tiles and clone the layout
  const {newLayout} = fixLayout(data.dashboard, newBPLayout)
  data.currentLayout = newLayout
}


// Grid item has moved. Save the updated layout
const handleItemMoved = (item:{}) => {
  console.log("DashboardView.handleItemMoved: ", item)
  Save()
}

// Grid item has resized. Save the updated layout
const handleItemResized = (item:{}) => {
  console.log("DashboardView.handleItemResized: ", item)
  Save()
}


// Save the layouts to the dashboard
// FIXME: when to save?
// Options:
//  1. In updateLayout, but not during initialization
//  2. Detect that the layout is modified instead of selected due to breakpoint or initialization
//     only need to save when items are add/removed/resized/moved
const Save = () => {
  if (data.dashboard) {
    let newDash = {...data.dashboard}
    // Make sure the current layout is saved as well
    let lastBreakpoint = gridLayout.value.lastBreakpoint
    gridLayout.value.layouts[lastBreakpoint] = data.currentLayout
    newDash.layouts = gridLayout.value.layouts
    console.info("DashboardView.Save: saving layouts", newDash.layouts)
    dashStore.UpdateDashboard(newDash)
    // As the dashboard is replaced in the store, changes to the existing 
    // dashboard won't be detected. Manual update instead.
    data.dashboard = dashStore.GetDashboardByName(props.dashboardName)
  }
}

// Ensure that the given layout contains all the dashboard tiles and remove unknown items
// Returns repaired layout and the number of changes made.
const fixLayout = (dashboard: DashboardDefinition|undefined, 
        currentLayout: Array<ILayoutItem>):
    {newLayout: Array<ILayoutItem>, changeCount: number} => {

  let newLayout = new Array<ILayoutItem>()
  let changeCount = 0
  if (!dashboard) {
    return {newLayout, changeCount}
  }
  // make sure all tiles are represented and remove old layout items for which there are no tiles
  let count=0, newCount = 0

  // Re-use existing tile layout and create new ones
  console.log("fixLayout: Finding items in currentLayout: ", currentLayout)
  for (let id in dashboard.tiles) {
    let item = currentLayout.find( (item)=>{
      if (!item) {
        console.warn("fixLayout. CurrentLayout contains unexpected null item. Ignored")
        return false
      }
      return (item.i === id)
    })
    count++
    if (!item) {
      // The tile doesn't have a layout item, add one
      newLayout.push({i:id, x:0, y:newCount, w:3, h:3})
      newCount++
    } else {
      // The tile already has a layout item, keep it
      newLayout.push(item)
    }
  }
  console.log("DashboardView.fixLayout for dashboard '",dashboard.name,"'.", count, "items of which",newCount,"are new.",
      "\nNew layout: ",newLayout)

  let removedCount = currentLayout.length - (newLayout.length-newCount)
  changeCount = (newCount+removedCount)
  return {newLayout, changeCount}
}

// // Provide the dashboard tile with the given ID and warn of unexpected errors
// const getTileByID = (id: string) : DashboardTileConfig|undefined => {
//   if (!data.dashboard || !data.dashboard.tiles) {
//     console.warn("DashboardView:getTileByID No Dashboard or missing dashboard tiles")
//     return undefined
//   }
//   let tile = data.dashboard.tiles[id]
//   if (!tile) {
//     console.warn("DashboardView:getTileByID can't find tile with ID", id)
//   }
//   return tile
// }



</script>


<template>
  <div v-if="!data.dashboard">
    <h4>Oops, dashboard {{props.dashboardName}} is not found. </h4>
  </div>
  <div v-else>
    <h4>Dashboard name={{props.dashboardName}} </h4>

    <!-- draggableCancel is used to prevent interference with tile menu click -->
    <GridLayout ref="gridLayout"
                :cols='{lg:24, md:16, sm:12, xs:8, xxs:4}'
                :layout="data.currentLayout"
                :responsiveLayouts="data.dashboard.layouts"
                :row-height="40"
                :verticalCompact="true"
        :is-draggable="appState.State().editMode"
        :is-resizable="appState.State().editMode"
        :responsive="true"
        :useCSSTransforms="true"
        @breakpoint-changed="handleBreakpointChange"
    >
      <grid-item v-for="item in data.currentLayout"
                 :i="item.i"
                 :x="item.x"
                 :y="item.y"
                 :w="item.w"
                 :h="item.h"
                 :minW="2"
                 :minH="2"
                 drag-ignore-from=".no-drag-area"
                 @moved="handleItemMoved"
                 @resized="handleItemResized"
      >
       <DashboardTile 
          :tile="data.dashboard?.tiles?.[item.i]"
          :dashboard="data.dashboard"
          :thingStore="thingStore"
          :dashStore="dashStore"
          />
      </grid-item> 
    </GridLayout>

  </div>
</template>
