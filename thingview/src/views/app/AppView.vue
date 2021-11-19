<script lang="ts" setup>
import {reactive} from "vue";
import AppHeader from "./AppHeader.vue";
// import {RouterView} from 'vue-router';

const appState = reactive({
  editMode: false,
  pages: [
      {label:'Page1', to:'/pages/page1'},
    {label:'Page2', to: '/pages/page2'}],
  selectedPage: 0,
});

const handleAddPage = (name:string) => {
  appState.pages.push({label:name, to:name});
  console.log("Added page: ",name)
}

const handleSelectPage = (index:number) => {
  console.log("Selecting page: ", index)
  appState.selectedPage = index;
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