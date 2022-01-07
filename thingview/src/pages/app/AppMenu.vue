<script lang="ts" setup>

import  {MenuAbout, MenuEditMode, MenuAddDashboard, MenuAccounts, MenuThings} from "@/pages/app/MenuConstants";

import TMenuButton, {IMenuItem} from "@/components/TMenuButton.vue";
import {
  matDashboard,
  matMenu,
  matInfo,
  matLensBlur,
  matLink,
  matAdd,
  matCheckBox,
  matCheckBoxOutlineBlank
} from "@quasar/extras/material-icons";

import { DashboardPrefix, AccountsRouteName, ThingsRouteName } from "@/router";
import {DashboardDefinition} from "@/data/dashboard/DashboardStore";

interface IAppMenu {
  editMode: boolean;
  dashboards: ReadonlyArray<DashboardDefinition>;
}
const props = withDefaults(defineProps<IAppMenu>(), {
      editMode: false,
      pages: ()=>[],
    })

const emit = defineEmits<{
  (e: 'onMenuAction', item:IMenuItem):void // IMenuItem
}>()

const handleMenuAction = (item:IMenuItem) => {
  console.debug("AppMenu.handleMenuAction. label=",item.label)
  emit('onMenuAction', item)
}

const getMenuItems = (dashboards: ReadonlyArray<DashboardDefinition>, editMode:boolean): Array<IMenuItem> => {
  let items:IMenuItem[] = dashboards.map(dash=>{
    return {
      id: dash.id,
      label: dash.name,
      icon: matDashboard,
      to: DashboardPrefix+"/"+dash.name
    }
  })
  items.push({separator:true})
  items.push({
    label: "Add Dashboard...",
    icon: matAdd,
    id: MenuAddDashboard,
  }, {
    label: "All Things...",
    icon: matLensBlur,
    id: MenuThings,
    to: {name:ThingsRouteName},
    // to: "/things"
  }, {
    label: "Edit Mode",
    icon: editMode ? matCheckBox : matCheckBoxOutlineBlank,
    id: MenuEditMode,
  }, {
    id: MenuAccounts,
    label: "Accounts...",
    icon: matLink,
    to: {name: AccountsRouteName},
  },{
    id: MenuAbout,
    label: "About...",
    icon: matInfo,
  })
  return items
}



</script>


<template>
  <TMenuButton 
    :icon="matMenu"
    :items="getMenuItems(props.dashboards, props.editMode)"
    @onMenuAction="handleMenuAction"
  />

</template>

