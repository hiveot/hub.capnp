<script lang="ts">
import {defineComponent} from "vue";
import AppHeader from "./AppHeader.vue";
import {RouterView} from 'vue-router';

export default defineComponent({
  components: { AppHeader, RouterView},
  
  emits: [
    "onEditModeChange", 
    "onPageSelect", 
  ],
  props: {
    editMode: Boolean,
    pages: Array,
    selectedPage: String,
  },
  setup(props,{emit}) {
    return {emit};
  }
})
</script>


<template>
<div style="display:flex; flex-direction: column;">
  <AppHeader :editMode="editMode" :pages="pages" :selectedPage="selectedPage" 
  @onEditModeChange='$emit("onEditModeChange",$event)'
  @onPageSelect='$emit("onPageSelect",$event)'
  />
  <RouterView></RouterView>
</div>
</template>
