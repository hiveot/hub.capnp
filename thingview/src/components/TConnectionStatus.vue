<script lang="ts" setup>
import {QBtn, QTooltip} from 'quasar';
// import TButton from './TButton.vue';
import {IConnectionStatus} from "@/data/ConnectionManager";
import {matLink, matLinkOff} from "@quasar/extras/material-icons";


const props = defineProps<{
  value: IConnectionStatus,
  to?: String|Object,
  withText?:boolean,
}>()

</script>

<template>
  <QBtn  flat
         :icon="(props.value.connected) ? matLink : matLinkOff"
         :text-color="(props.value.authenticated) ? 'green' : 'red'"
         :to="to"
  >
    <QTooltip  class="text-body2" transition-show="scale">
      {{value.statusMessage}}
    </QTooltip>
  </QBtn>
  <div v-if="props.withText">{{value.statusMessage}}</div>

</template>