
<script lang="ts" setup>
import {reactive} from "vue";
import AppMenu, { MenuAbout, MenuEditMode, MenuAddPage} from './AppMenu.vue';

import AboutDialog from "./AppAboutDialog.vue";
import AddPageDialog from "./AppAddPageDialog.vue";

import TabView from 'primevue/tabview';
import TabMenu from 'primevue/tabmenu';
import InputSwitch from 'primevue/inputswitch';
import IconLink from '~icons/mdi/link'


interface IAppHeader {
  editMode: boolean
  pages: Array<{ label:string, icon?:string, to?:string }>
  selectedPage: number
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

const handleAddPage = (name:string) => {
  console.log("Adding page", name);
  emit("onAddPage", name);
}
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
const handleTabSelect = (event:any) => {
  console.log("emit onPageSelect: page=",event.index)
  emit("onPageSelect", event.index);
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
    <TabMenu :model="props.pages"
        :activeIndex="props.selectedPage"
        @tab-click='handleTabSelect'
    />

    <div style="flex-grow:1"/>

    <!-- Edit mode switching -->
    <div style="display: flex; flex-direction: row; align-items: center">
      <InputSwitch :modelValue="props.editMode"
                @input="handleEditModeChange"
                inactive-color="gray"
      />
      <label style="padding-left:5px; color:dimgray">Edit</label>
    </div>

    <!-- Connection Status -->
    <Button class="p-button-text"
            v-tooltip="'Connection Status & Configuration'"
    >
      <IconLink/>
    </Button>

    <!-- Dropdown menu -->
    <AppMenu :pages="props.pages"  :editMode="props.editMode"
             @onMenuSelect="handleMenuSelect"
    />

  </div>
</template>

<style>
/* Tab bar should have header background */
.p-tabmenu .p-tabmenu-nav {
  background: transparent !important;
}
.p-tabmenu .p-tabmenu-nav .p-tabmenuitem .p-menuitem-link {
  background:transparent !important;
}
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

