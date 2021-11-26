<script lang="ts" setup>
// import { useQuasar } from 'quasar'

import AppHeader from "./AppHeader.vue";
import {IMenuItem} from "@/components/MenuButton.vue";

import appState from '@/data/AppState'


// future option for dark theme setting
// const $q = useQuasar()
// $q.dark.set(true) // or false or "auto"
// $q.dark.toggle()

</script>


<template>
<div class="appView">
  <AppHeader  :appState="appState" />
  <router-view></router-view>
</div>
</template>


<style>
.appView {
  display:flex;
  flex-direction: column;
}
</style>