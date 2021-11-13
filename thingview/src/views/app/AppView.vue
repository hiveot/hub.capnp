<script lang="ts">
import {defineComponent, reactive} from "vue";
import AppHeader from "./AppHeader.vue";
import {RouterView} from 'vue-router';

export default defineComponent({
  components: { AppHeader, RouterView},
  
  setup(props,{emit}) {

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

    return {emit, appState, handleAddPage};
  }
})
</script>


<template>
<div class="appView">
  <AppHeader
      :editMode="appState.editMode"
      :pages="appState.pages"
      :selectedPage="appState.selectedPage"
       @onEditModeChange="appState.editMode = $event"
       @onPageSelect="appState.selectedPage = $event"
      @onAddPage="handleAddPage"
  />
  <RouterView></RouterView>
</div>
</template>


<style>
.appView {
  display:flex;
  flex-direction: column;
}
</style>