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
    "onEditModeChange", 
    "onPageSelect", 
    "onMenuSelect"
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
    const handleAddPageClosed = () => {
      // console.log("About closed...");
      data.showAddPage = false;
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
      handleOpenAddPage,  handleAddPageClosed,
      handleEditModeChange,  
      handleMenuSelect, handleTabSelect};
  }
})
</script>


<template>
<div style="display:flex; flex-direction: row; 
align-items: center; justify-content: flex-start; gap:10px;
height:42px; background-color: rgb(218, 229, 231);">

  <AboutDialog :visible="data.showAbout" @onClosed='handleAboutClosed'/>
  <AddPageDialog :visible="data.showAddPage" @onClosed='handleAddPageClosed'/>

  <img alt="logo" src="@/assets/logo.png" @click="handleOpenAbout" style="height: 26px; height:100%; cursor:pointer; padding-right:5px;"/>
  
  <!-- On larger screens show a tab bar for dashboard pages -->
  <ElTabs style="align-self:bottom" 
    :model-value="selectedPage" 
    @tab-click='$emit("onPageSelect", $event.paneName)'
  >
    <ElTabPane v-for="page in pages" :label="page" :key="page" :name="page"/>
  </ElTabs>
  
  <!-- Edit mode switching -->
  <div style="flex-grow:1"/>
    <ElSwitch :value="editMode"  @change="handleEditModeChange" active-text="Edit" 
       inactive-color="gray"
    />

  <!-- Connection Status -->
  <button className="buttonHover">
    <!-- <v-icon name="md-link" scale="1" /> -->
    <icon icon="mdi:link-off" height="24"  />
  </button>
  
  <!-- Dropdown menu -->
  <AppMenu :pages="pages"  :editMode="editMode" 
    @onMenuSelect="handleMenuSelect" 
  />

</div>
</template>

<style>
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