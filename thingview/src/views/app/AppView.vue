<script lang="ts" setup>
import {reactive} from "vue";
import AppHeader from "./AppHeader.vue";
// import {RouterView} from 'vue-router';

const appState = reactive({
  editMode: false,
  pages: ['Page1', 'Page2'],
  selectedPage: "Page1",
});

const handleAddPage = (name:string) => {
  appState.pages.push(name);
  console.log("Added page: ",name)
}

const handleSelectPage = (event:any) => {
  console.log("Selecting page: ", event)
  appState.selectedPage = event;
}


</script>


<template>
<div class="appView">
  <AppHeader
      :editMode="appState.editMode"
      :pages="appState.pages"
      :selectedPage="appState.selectedPage"
       @onEditModeChange="appState.editMode = $event"
       @onPageSelect="handleSelectPage"
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