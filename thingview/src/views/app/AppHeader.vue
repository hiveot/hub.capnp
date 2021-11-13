
<script lang="ts">
import {defineComponent, reactive} from "vue";
import { ElTabs, ElTabPane, ElSwitch } from "element-plus";
import AppMenu, {MenuAbout, MenuEditMode, MenuAddPage} from './AppMenu.vue';

import { addIcon } from '@iconify/vue/dist/offline';
import  MdiLink  from '@iconify/icons-mdi/link';
import  MdiLinkOff  from '@iconify/icons-mdi/link-off';
import AboutDialog from "./AppAboutDialog.vue";
import AddPageDialog from "./AppAddPageDialog.vue";
// import DashboardView from "../DashboardView.vue";


export default defineComponent({
  components: { AboutDialog, AddPageDialog, AppMenu, ElTabs, ElTabPane, ElSwitch,},
  props: {
    editMode: Boolean,
    pages: Array,
    selectedPage: String,
  },
  
  emits: [
    "onAddPage",
    "onEditModeChange",
    "onPageSelect", 
    "onMenuSelect",
  ],

  setup(props, {emit}) {
    addIcon('mdi:link', MdiLink);
    addIcon('mdi:link-off', MdiLinkOff);

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
    return {data, 
      handleAboutClosed,  handleOpenAbout, 
      handleOpenAddPage,  handleAddPage,
      handleEditModeChange,  
      handleMenuSelect, handleTabSelect};
  }
})
</script>

<template>
  <div class="header">

    <AboutDialog :visible="data.showAbout" @onClosed='handleAboutClosed'/>
    <AddPageDialog :visible="data.showAddPage"  @onClosed='data.showAddPage = false;' @onAdd="handleAddPage"/>

    <img alt="logo" src="@/assets/logo.png" @click="handleOpenAbout"
         style="height: 32px;cursor:pointer; padding:5px;"
    />

    <!-- On larger screens show a tab bar for dashboard pages -->
    <ElTabs
        :model-value="selectedPage"
        @tab-click='handleTabSelect'
    >
      <ElTabPane v-for="page in pages" :name="page" :key="page" :label="page"/>
    </ElTabs>

    <!-- Edit mode switching -->
    <div style="flex-grow:1"/>
    <ElSwitch :value="editMode"  @change="handleEditModeChange" active-text="Edit"
              inactive-color="gray"
    />

    <!-- Connection Status -->
    <button className="buttonHover">
      <!-- <v-icon name="md-link" scale="1" /> -->
      <v-icon icon="mdi:link-off" height="24"  />
    </button>

    <!-- Dropdown menu -->
    <AppMenu :pages="pages"  :editMode="editMode"
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
  font-size: large;
  gap: 10px;
  height: 42px;
  background-color: rgb(218, 229, 231);
}
.buttonHover {
  background-color: transparent;
  border: 0px;
  color: rgb(78, 77, 77);
}
.buttonHover:hover {
  background-color: rgb(235, 231, 231);
}
/* drop the tab select indicator to the bottom  */
.el-tabs__header {
  margin: 0;
}
</style>

