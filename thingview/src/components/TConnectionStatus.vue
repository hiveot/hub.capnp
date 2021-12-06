<script lang="ts" setup>
import {QBtn, QTooltip} from 'quasar';

import {IConnectionStatus} from "@/data/ConnectionManager";
import {matLink, matLinkOff} from "@quasar/extras/material-icons";


const props = defineProps<{
  value: IConnectionStatus,
  to?: String|Object,
  tooltip?: String,
}>()

</script>

<template>
  <QBtn  flat
         :icon="(props.value.connected) ? matLink : matLinkOff"
         :color="(props.value.authenticated) ? 'green' : 'red'"
         :to="to"
  >
    <QTooltip>Connection Status & Configuration</QTooltip>
  </QBtn>

</template>