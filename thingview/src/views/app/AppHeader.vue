
<script lang="ts" setup>
import {reactive} from "vue";
import AppMenu, { MenuAbout, MenuEditMode, MenuAddPage} from './AppMenu.vue';

import AboutDialog from "./AppAboutDialog.vue";
import AddPageDialog from "./AppAddPageDialog.vue";
// import DashboardView from "../DashboardView.vue";


interface IAppHeader {
  editMode: boolean
  pages: Array<string>
  selectedPage: string
}
const props = defineProps<IAppHeader>()
  
const emit = defineEmits([
    "onAddPage",
    "onEditModeChange",
    "onPageSelect", 
    "onMenuSelect",
  ])

const data =reactive({
  showAbout: false,
  showAddPage: false,
})

const handleEditModeChange = (ev:any)=>{
  console.log("AppHeader: emit onEditModeChange")
  emit("onEditModeChange", ev);
}
const handleOpenAbout = () => {
  // console.log("Opening about...");
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
const handleAddPage = (name:string) => {
  console.log("Adding page", name);
  emit("onAddPage", name);
}
const handleTabSelect = (tabPane:any) => {
  console.log("emit onPageSelect: page=",tabPane.paneName)
  emit("onPageSelect", tabPane.paneName);
}
const handleMenuSelect = (menu:string) => {
  console.log("handleMenuSelect: ", menu);
  if (menu == MenuAbout) {
    handleOpenAbout();
  } else if (menu == MenuEditMode) {
    handleEditModeChange(!props.editMode);
  } else if (menu == MenuAddPage) {
    handleOpenAddPage();
  } else {
    console.log("emit onPageSelect: page=",menu)
    emit("onPageSelect", menu);
  }
}
console.log("AppHeader: selectedPage="+props.selectedPage);


</script>

<template>
  <div class="header">

    <AboutDialog :visible="data.showAbout" @onClosed='handleAboutClosed'/>
    <AddPageDialog :visible="data.showAddPage"  @onClosed='data.showAddPage = false;' @onAdd="handleAddPage"/>

    <img alt="logo" src="@/assets/logo.png" @click="handleOpenAbout"
         style="height: 40px;cursor:pointer; padding:5px;"
    />

    <!-- On larger screens show a tab bar for dashboard pages -->
    <q-tabs
        :model-value="props.selectedPage"
        @tab-click='handleTabSelect'
    >
      <q-tab v-for="page in props.pages" :name="page"  :label="page"/>
    </q-tabs>

    <!-- Edit mode switching -->
    <div style="flex-grow:1"/>
    <q-toggle :model-value="props.editMode"
              @update:model-value="handleEditModeChange"
              label="Edit"
              inactive-color="gray"
    />

    <!-- Connection Status -->
    <q-btn flat icon="mdi-link-off">
      <q-tooltip>Connection Status & Configuration</q-tooltip>
    </q-btn>

    <!-- Dropdown menu -->
    <AppMenu :pages="props.pages"  :editMode="props.editMode"
             @onMenuSelect="handleMenuSelect"
    />

  </div>
</template>

<style>
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

