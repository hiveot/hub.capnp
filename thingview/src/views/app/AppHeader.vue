
<script lang="ts" setup>
import {reactive} from "vue";
import AppMenu, { MenuAbout, MenuEditMode, MenuAddPage} from './AppMenu.vue';

import AboutDialog from "./AppAboutDialog.vue";
import AddPageDialog from "./AppAddPageDialog.vue";
import {IMenuItem} from "@/components/MenuButton.vue";

import {mdiLink, mdiLinkOff} from "@quasar/extras/mdi-v6";

interface IAppHeader {
  editMode: boolean
  pages: Array<IMenuItem>
}
const props = defineProps<IAppHeader>()

const emit = defineEmits([
    "onAddPage",
    "onEditModeChange",
    "onMenuSelect",
  ])

const data =reactive({
  showAbout: false,
  showAddPage: false,
})

const handleAddPage = (name:string) => {
  console.log("Adding page", name);
  emit("onAddPage", name);
}
const handleEditModeChange = (ev:any)=>{
  console.log("AppHeader: emit onEditModeChange")
  emit("onEditModeChange", ev);
}
const handleOpenAbout = () => {
  console.log("Opening about...");
  data.showAbout = !data.showAbout;
}
const handleOpenAddPage = () => {
  // console.log("Opening about...");
  data.showAddPage = !data.showAddPage;
}
const handleAboutClosed = () => {
  // console.log("About closed...");
  data.showAbout = false;
}
// handle Dialog and edit mode select
const handleMenuSelect = (menuItem:IMenuItem) => {
  console.log("handleMenuSelect: ", menuItem);
  if (menuItem.id == MenuAbout) {
    handleOpenAbout();
  } else if (menuItem.id == MenuEditMode) {
    handleEditModeChange(!props.editMode);
  } else if (menuItem.id == MenuAddPage) {
    handleOpenAddPage();
  } else {
  }
}

</script>

<template>
  <div class="header">

    <AboutDialog :visible="data.showAbout"
                 @onClosed='handleAboutClosed'/>
    <AddPageDialog :visible="data.showAddPage"
                   @onClosed='data.showAddPage=false'
                   @onAdd="handleAddPage"/>

    <img alt="logo" src="@/assets/logo.png" @click="handleOpenAbout"
         style="height: 40px;cursor:pointer; padding:5px;"
    />

    <!-- On larger screens show a tab bar for dashboard pages -->
    <q-tabs   inline-label indicator-color="green">
      <q-route-tab v-for="page in props.pages"
             :label="page.label"
             :icon="page.icon"
             :to="(page.to == undefined) ? '' : page.to"
      />
    </q-tabs>

    <div style="flex-grow:1"/>

    <!-- Edit mode switching -->
    <q-toggle :model-value="props.editMode"
              @update:model-value="handleEditModeChange"
              label="Edit"
              inactive-color="gray"
    />

    <!-- Connection Status -->
<!--    <Button  icon="mdi-link-off" flat tooltip="Connection Status & Configuration"/>-->
    <Button  :icon="mdiLinkOff" flat tooltip="Connection Status & Configuration"/>

    <!-- Dropdown menu -->
    <AppMenu :pages="props.pages"
             :editMode="props.editMode"
             @onMenuSelect="handleMenuSelect"
    />

  </div>
</template>

<style>
/* Tab bar should have header background */
/*.p-tabmenu .p-tabmenu-nav {*/
/*  background: transparent !important;*/
/*}*/
/*.p-tabmenu .p-tabmenu-nav .p-tabmenuitem .p-menuitem-link {*/
/*  background:transparent !important;*/
/*}*/

.header {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: flex-start;
  /*font-size: large;*/
  gap: 10px;
  /*height: 46px;*/
  background-color: rgb(218, 229, 231);
}
</style>

