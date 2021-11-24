<script lang="ts" setup>
import {reactive} from "vue";
import AppHeader from "./app/AppHeader.vue";
import {IMenuItem} from "@/components/MenuButton.vue";
import { useQuasar } from 'quasar'

import appState, {PagesPrefix} from '@/store/AppState'


// future option for dark theme setting
const $q = useQuasar()
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