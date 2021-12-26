<script lang="ts" setup>
import {nextTick, onMounted, reactive, ref, watch} from "vue";
import {GridLayout, GridItem} from 'vue3-grid-layout'

import ds, {DashboardDefinition, DashboardTile } from '@/data/dashboard/DashboardStore'
import appState from '@/data/AppState'
import TGridTile from "@/components/TGridTile.vue";


// Dashboard view shows the dashboard with the given name 
const props = defineProps<{
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
  dashboard: ds.GetDashboardByName(props.dashboardName),
  currentLayout:  Array<ILayoutItem>(),
  initializing: false,
})

watch(()=>props.dashboardName, (newName:string)=>{
  console.log("DashboardView.watch")
  data.dashboard = ds.GetDashboardByName(props.dashboardName)
  // a new dashboard needs restoring the layouts of all breakpoints
  if (data.dashboard) {
    handleNewDashboard(data.dashboard)
  }
})

onMounted(()=>{
  nextTick(()=>{

  console.log("DashboardView.onMounted")
  data.dashboard = ds.GetDashboardByName(props.dashboardName)

  // a new dashboard needs restoring the layouts of all breakpoints
  if (data.dashboard) {
    handleNewDashboard(data.dashboard)
  }
  })
})

// After changing the dashboard this updates the layouts for all breakpoints
// FIXME: this results an a layout update that saves the new layout back to the dashboard .. twice
// Options:
//  1. Detect in updateLayout that initialization is in progress and don't save
//  2. Detect that the layout is modified instead of changed (due to breakpoint changes)
//     only need to save when items are add/removed/resized/moved
//
// FIXME: when dashboard changes re-initialize. This doesn't seem to work.
const handleNewDashboard = (dashboard:DashboardDefinition) => {
  data.initializing = true
  // a new dashboard needs restoring the layouts of all breakpoints
  if (!dashboard) {
    return
  } else if (!dashboard.layouts) {
    console.error("handleNewDashboard: no layouts")
  }
  console.info("DashboardView.handleNewDashboard. Start. Restoring all layouts from saved dashboard",
      dashboard.name, ". \nlayouts: ", dashboard.layouts)

  // reset the layout and trigger an update of the current layout
  // gridLayout.value.responsiveLayouts = dashboard.layouts
  gridLayout.value.lastBreakpoint = null
  // debugger
  gridLayout.value.initResponsiveFeatures()
  gridLayout.value.responsiveGridLayout()

  console.log("DashboardView.handleNewDashboard. Done")
  data.initializing = false
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
  
  let lastBreakpoint:string = gridLayout.value.lastBreakpoint
  if (lastBreakpoint ) {
    // Workaround: GridLayout hasn't saved the previous layout
    gridLayout.value.layouts[lastBreakpoint] = [...data.currentLayout]

    // Make sure the new layout contain all tiles and clone the layout
    const {newLayout, changeCount} = fixLayout(data.dashboard, newBPLayout)
    data.currentLayout = newLayout
  }
  Save()
}

// dashboard layout has changed.
const handleLayoutUpdated = (newLayout:any) => {
  console.log("DashboardView.handleLayoutUpdated: newLayout", newLayout)
  // console.log("handleLayoutUpdate: newLayout", newLayout)
  // data.currentLayout = newLayout
  let lastBreakpoint = gridLayout.value.lastBreakpoint
  gridLayout.value.layouts[lastBreakpoint] = data.currentLayout
}

// Save the layouts to the dashboard
const Save = () => {
  if (data.dashboard) {
    let newDash = {...data.dashboard}
    // Make sure the current layout is saved as well
    // let lastBreakpoint = gridLayout.value.lastBreakpoint
    // gridLayout.value.layouts[lastBreakpoint] = data.currentLayout
    newDash.layouts = gridLayout.value.layouts
    console.info("DashboardView.Save: saving layouts", newDash.layouts)
    ds.UpdateDashboard(newDash)
  }
}

// Ensure that the given layout contains all the dashboard tiles and remove unknown items
// Returns repaired layout and the number of changes made.
const fixLayout = (dashboard: DashboardDefinition|undefined, currentLayout: Array<ILayoutItem>):
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
        console.error("fixLayout. CurrentLayout contains null item. Ignored")
        return false
      }
      return (item.i === id)
    })
    count++
    if (!item) {
      // The tile doesn't have a layout item, add one
      newLayout.push({i:id, x:0, y:newCount, w:1, h:1})
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


</script>


<template>
  <div v-if="!data.dashboard">
    <h4>Oops, this dashboard is not found.</h4>
  </div>
  <div v-else>
    <h4>Dashboard name={{data.dashboard.name}} </h4>

    <!-- draggableCancel is used to prevent interference with tile menu click -->
    <GridLayout ref="gridLayout"
                :layout="data.currentLayout"
                :responsiveLayouts="data.dashboard.layouts"
        :col-num="12"
        :row-height="40"
        :is-draggable="appState.State().editMode"
        :is-resizable="appState.State().editMode"
        :responsive="true"
        :useCSSTransforms="true"
        draggableCancel=".not-draggable-area"
        @layout-updated="handleLayoutUpdated"
        @breakpoint-changed="handleBreakpointChange"
    >
      <grid-item v-for="item in data.currentLayout"
                 :i="item.i"
                 :x="item.x"
                 :y="item.y"
                 :w="item.w"
                 :h="item.h"
      >
        <TGridTile :id="item.i" :tile="data.dashboard.tiles[item.i]"/>
      </grid-item>
    </GridLayout>

  </div>
</template>
