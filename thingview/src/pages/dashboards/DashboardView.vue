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
// 4. There is no easy way to detect if a widget has moved or resized, so the changes to the
//    layout can be saved.
// 3. There is no easy way to switch dashboards and update the layout
//    Ideally you bind a new responsive layout and done. That doesn't work though.

import {nextTick, onMounted, reactive, ref, watch} from "vue";
import {GridLayout, GridItem} from 'vue3-grid-layout'

import ds, {DashboardDefinition } from '@/data/dashboard/DashboardStore'
import appState from '@/data/AppState'
import DashboardWidget from "./DashboardWidget.vue";


/**
 * Dashboard view shows the dashboard with the given name
 */
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

watch(()=>props.dashboardName, ()=>{
  console.log("DashboardView.watch(dashboardName)")
  let dashboard = ds.GetDashboardByName(props.dashboardName)
  if (dashboard) {
    handleNewDashboard(dashboard)
  }
})

onMounted(()=>{
  console.log("DashboardView.onMounted")
  let dashboard = ds.GetDashboardByName(props.dashboardName)
  if (dashboard) {
    handleNewDashboard(dashboard)
  }
})

// After changing the dashboard this updates the layouts for all breakpoints
const handleNewDashboard = (dashboard:DashboardDefinition) => {
  // a new dashboard needs restoring the layouts of all breakpoints
  data.dashboard = ds.GetDashboardByName(props.dashboardName)
  if (!dashboard.layouts) {
    console.error("handleNewDashboard: no layouts. Ignored")
    return
  } else if (!gridLayout || !gridLayout.value) {
    console.warn('handleNewDashboard: ref gridLayout not initialized')
    return
  }
  console.info("DashboardView.handleNewDashboard. Start. Restoring all layouts from saved dashboard",
      dashboard.name, ". \nlayouts: ", dashboard.layouts)

  data.initializing = true

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
    data.initializing = false
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

// dashboard layout has changed.
// This event is received when:
//  1. dragging ended -> save layout
//  2. resizing ended -> save layout
//  3. gridlayout was mounted -> don't save layout as nothing changed
//  4. method layoutUpdate() was invoked and differs from original layout -> don't save layout
const handleLayoutUpdated = (newLayout:[]) => {
  // if newLayout is empty then there is nothing to save
  if (!data.dashboard) {
    console.log("DashboardView.handleLayoutUpdated: Missing dashboard. Update ignored.")
    return
  } else if (!newLayout || newLayout.length == 0) {
    console.log("DashboardView.handleLayoutUpdated: newLayout is empty. Update ignored.")
    return
  } else if (data.initializing) {
    console.log("DashboardView.handleLayoutUpdated: initializing. Update ignored.")
    return;
  }
  let breakpoint = gridLayout.value.lastBreakpoint
  let oldLayout:{}[] = gridLayout.value.layouts[breakpoint]
  if (!oldLayout || (newLayout.length != oldLayout.length)) {
    console.info("DashboardView.handleLayoutUpdated: old layout differs from new. Saving...: oldLayout:", oldLayout, " newLayout:", newLayout)
    Save()
    return
  }
  let misMatch = newLayout.find( (newItem:any) => {
    let oldItem:any = oldLayout.find( (item:any) => item.i == newItem.i)
    if (!oldItem) {
      return true // there is a mismatch
    }
    return (oldItem.x != newItem.x || oldItem.y != newItem.y || oldItem.w != newItem.w || oldItem.h != newItem.h)
  })
  if (misMatch) {
    console.info("DashboardView.handleLayoutUpdated: Layout has differences. Saving")
    Save()
  } else {
    console.log("DashboardView.handleLayoutUpdated: Layout has no differences. Update ignored. oldLayout:", oldLayout, "newLayout:", newLayout)
  }
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
        console.warn("fixLayout. CurrentLayout contains unexpected null item. Ignored")
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
    <h4>Oops, dashboard {{props.dashboardName}} is not found. </h4>
  </div>
  <div v-else>
    <h4>Dashboard name={{props.dashboardName}} </h4>

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
        <DashboardWidget :id="item.i" :tile="data.dashboard.tiles[item.i]"/>
      </grid-item>
    </GridLayout>

  </div>
</template>
