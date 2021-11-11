<script  lang="ts">
import {defineComponent, reactive} from "vue";
import AppView from '@/views/app/AppView.vue';

export default defineComponent({
  components: {AppView},

  setup() {
    const appState = reactive({
      editMode: false,
      pages: ['page42'],
      selectedPage: 'page42',
    });
    return {appState};
  }
})

</script>

<template>
  <AppView 
    :editMode="appState.editMode"
    :selectedPage="appState.selectedPage"
    @onEditModeChange="appState.editMode = $event" 
    @onPageSelect="appState.selectedPage = $event" 
  />
</template>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: darkblue;
  margin-top: 60px;
}
</style>
