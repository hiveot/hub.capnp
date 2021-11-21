<script lang="ts" setup>
import {reactive} from "vue";
import AppHeader from "./AppHeader.vue";
import {IMenuItem} from "@/components/MenuButton.vue";

// import {RouterView} from 'vue-router';
import {PagesPrefix} from "@/router";
import {mdiViewDashboard} from "@quasar/extras/mdi-v6";

const appState = reactive({
  editMode: false,
  pages:[
    <IMenuItem>{label:'Page1', to: PagesPrefix+'/page1', icon:mdiViewDashboard},
    <IMenuItem>{label:'Page2', to: PagesPrefix+'/page2', icon:mdiViewDashboard}],
});

const handleAddPage = (name:string) => {
  appState.pages.push({label:name, to:PagesPrefix+'/'+name,icon:"mdi-view-dashboard"});
  console.log("Added page: ",name)
}
</script>


<template>
<div class="appView">
  <AppHeader
      :editMode="appState.editMode"
      :pages="appState.pages"
       @onEditModeChange="appState.editMode = $event"
      @onAddPage="handleAddPage"
  />
  <router-view></router-view>
</div>
</template>


<style>
.appView {
  display:flex;
  flex-direction: column;
}
</style>