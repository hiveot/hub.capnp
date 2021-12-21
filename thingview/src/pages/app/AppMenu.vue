<script lang="ts" setup>

import  {MenuAbout, MenuEditMode, MenuAddDashboard, MenuAccounts, MenuThings} from "@/pages/app/MenuConstants";

import TMenuButton, {IMenuItem} from "@/components/TMenuButton.vue";
import {
  matMenu,
  matInfo,
  matLensBlur,
  matLink,
  matAdd,
  matCheckBox,
  matCheckBoxOutlineBlank
} from "@quasar/extras/material-icons";

import { AccountsRouteName, ThingsRouteName } from "@/data/AppState";

interface IAppMenu {
  editMode: boolean;
  dashboards: Array<IMenuItem>;
}
const props = withDefaults(defineProps<IAppMenu>(), {
      editMode: false,
      pages: ()=>[],
    })

const emit = defineEmits<{
  (e: 'onMenuSelect', item:IMenuItem):void // IMenuItem
}>()

const handleMenuSelect = (item:IMenuItem) => {
  console.log('AppMenu: onMenuSelect: ', item);
  emit("onMenuSelect", item);
}


const getMenuItems = (dashboards: Array<IMenuItem>, editMode:boolean): Array<IMenuItem> => {
  return  [...dashboards, {
    separator: true,
  }, {
    label: "All Things...",
    icon: matLensBlur,
    id: MenuThings,
    to: {name:ThingsRouteName},
    // to: "/things"
  }, {
    label: "Add Dashboard...",
    icon: matAdd,
    id: MenuAddDashboard,
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
  }]
}



</script>


<template>
  <TMenuButton 
    :icon="matMenu"
    :items="getMenuItems(props.dashboards, props.editMode)"
    @onMenuSelect='handleMenuSelect'
  />

</template>

