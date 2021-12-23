<script lang="ts" setup>
import {reactive} from "vue";
import VueGridLayout from 'vue-grid-layout'

import {useQuasar} from "quasar";
import ds, {DashboardDefinition} from '@/data/dashboard/DashboardStore'
import router, {DashboardPrefix} from "@/router";

const $q = useQuasar()
//
// const NotifyMissingDashboard = ():string => {
//   $q.notify({
//     position: "top",
//     message:`Dashboard not found. Automatically redirected to first dashboard`,
//     type: "warning",
//   })
//
//   // redirect to first valid page
//   let dash = ds.dashboards[0]
//   if (!dash) {
//     // no dashboards defined
//     router.push('/')
//   } else {
//     // return DashboardPrefix + "/" + dash.name
//     router.push(DashboardPrefix + "/" + dash.name)
//   }
//   return "/"
// }

// Dashboard view shows the dashboard with the given name 
const props = defineProps<{
    dashboard?: DashboardDefinition,
    editMode?: boolean,
  }>()

const data = reactive({
})

</script>

<template>
<!--  <grid-layout -->
<!--      layout="layout"-->
<!--      :row-height="40"-->
<!--      :width="600"-->
<!--      :is-draggable="props.editMode"-->
<!--      :is-resizable="props.editMode"-->
<!--  >-->
<!--    <grid-item-->
<!--        i="" h="" w="" y="" x=""-->
<!--        />-->
<!--  </grid-layout>-->
  <div v-if="!props.dashboard">
    <h4>Oops, this dashboard is not found.</h4>
<!--    <a :href="NotifyMissingDashboard()">Return to first dashboard</a>-->
  </div>
  <div v-else>
    <h4>Dashboard name={{props.dashboard.name}} </h4>
  </div>
</template>
