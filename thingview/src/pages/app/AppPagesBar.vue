<script lang="ts" setup>

import {QTabs, QRouteTab} from 'quasar'
import { DashboardDefinition } from '@/data/dashboard/DashboardStore';
import {DashboardPrefix} from '@/router'
import { MenuAddTile,  MenuEditDashboard, MenuDeleteDashboard } from './MenuConstants';
import TMenuButton, { IMenuItem } from '@/components/TMenuButton.vue';

import {matMenu, matAdd, matEdit, matDelete} from "@quasar/extras/material-icons"
import { ref } from 'vue';
import { emitKeypressEvents } from 'readline';

const props = defineProps<{
  dashboards: ReadonlyArray<DashboardDefinition>
  editMode?: boolean
}>()

const emit=defineEmits<{
  (e:'onMenuAction', item:IMenuItem, dashboard:DashboardDefinition):void,
  (e:'onSelectedTab', value:string):void,
}>()

const selectedTab = ref("")

// Page tab bar dropdown menu items
const pageMenuItems: IMenuItem[] = [
  {id: MenuAddTile, label: 'Add Tile', icon: matAdd},
  {separator: true},
  {id: MenuEditDashboard, label: 'Edit Dashboard', icon: matEdit},
  {id: MenuDeleteDashboard, label: 'Delete Dashboard', icon: matDelete},
]

// Submit event if the selected tab changes
// This submits null if not tab is selected, eg a different view is shown
const selectedTabUpdated = (value:any)=>{
  console.log("AppPagesBar.selectedTabUpdated. New value: ", value)
  selectedTab.value = value
  emit('onSelectedTab', value)
}

</script>

<template>
    <!-- On larger screens show a tab bar for dashboard page -->
    <QTabs   inline-label indicator-color="red"
    :model-value="selectedTab"
     @update:modelValue="selectedTabUpdated"
    >
      <!-- <QRouteTab v-for="dashboard in props.dashboards"
             :label="dashboard.label"
             :icon="dashboard.icon"
             :to="(dashboard.to === undefined) ? '' : dashboard.to"
      > -->
      <QRouteTab v-for="dashboard in props.dashboards"
             :label="dashboard.name"
             :to="DashboardPrefix+'/'+dashboard.name"
      >
        <TMenuButton v-if="props.editMode" 
          class="q-pa-xs"
          :icon="matMenu" 
          :items="pageMenuItems" 
          @onMenuAction="(item)=>emit('onMenuAction', item, dashboard)"
        />
      </QRouteTab>
    </QTabs>


</template>