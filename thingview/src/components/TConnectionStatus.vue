<script lang="ts" setup>

import {QBtn, QTooltip} from 'quasar';
import {IConnectionStatus} from "@/data/accounts/ConnectionManager";
import {matLink, matLinkOff} from "@quasar/extras/material-icons";

/** Button showing the connection status with colored link icon */
const props = defineProps<{
  /** The connection status to show */
  value: IConnectionStatus,
  /** optional route to navigate to when clicking the button */
  to?: String|Object,
  /** optionally show text description of the connection status below the icon */
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