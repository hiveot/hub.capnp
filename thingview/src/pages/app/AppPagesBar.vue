<script lang="ts" setup>

import {QTabs, QRouteTab} from 'quasar'
import { IDashboardRecord } from '@/data/AppState';
import TMenuButton, { IMenuItem } from '@/components/TMenuButton.vue';
import {matMenu, matAdd, matEdit, matDelete} from "@quasar/extras/material-icons"
import { MenuAddTile, MenuAddDashboard, MenuEditDashboard, MenuDeleteDashboard } from './MenuConstants';

const props = defineProps<{
  dashboards: IDashboardRecord[]
  editMode?: boolean
}>()

const emit=defineEmits<{
  (e:'onMenuSelect', item:IMenuItem, dashboard:IDashboardRecord):void,
}>()

// Pass selected menu
const handleMenuSelect = (item:IMenuItem, dashboard:IDashboardRecord) => {
  console.log('AppPagesBar: onMenuSelect: id=%s, dashboard=%s', item.id, dashboard.label);
  emit("onMenuSelect", item, dashboard);
}

// Page tab bar dropdown menu items
const pageMenuItems: IMenuItem[] = [
  {id: MenuAddTile, label: 'Add Tile', icon: matAdd},
  {separator: true},
  {id: MenuEditDashboard, label: 'Edit Dashboard', icon: matEdit},
  // {id: MenuAddDashboard, label: 'Add Dashboard', icon: matAdd},
  {id: MenuDeleteDashboard, label: 'Delete Dashboard', icon: matDelete},
]

</script>

<template>
    <!-- On larger screens show a tab bar for dashboard page -->
    <QTabs   inline-label indicator-color="green">
      <!-- <QRouteTab v-for="dashboard in props.dashboards"
             :label="dashboard.label"
             :icon="dashboard.icon"
             :to="(dashboard.to === undefined) ? '' : dashboard.to"
      > -->
      <QRouteTab v-for="dashboard in props.dashboards"
             :label="dashboard.label"
             :to="(dashboard.to === undefined) ? '' : dashboard.to"
      >
        <TMenuButton v-if="props.editMode" 
          class="q-pa-xs"
          :icon="matMenu" 
          :items="pageMenuItems" 
          @on-menu-select="(item)=>handleMenuSelect(item, dashboard)"
        />
      </QRouteTab>
    </QTabs>


</template>