<script lang="ts" setup>

import {ref} from 'vue'
import {ThingTD, TDProperty, TDAction, TDEvent} from '@/data/td/ThingTD';
import { QCard, QTab, QTabs, QTabPanel, QTabPanels } from 'quasar';
import {matSettings, matSettingsRemote, matDescription, matDirectionsRun} from '@quasar/extras/material-icons'

import ThingEvents from './ThingEvents.vue'
import ThingActions from "@/pages/things/ThingActions.vue";
import ThingAttributes from "@/pages/things/ThingAttributes.vue";
import ThingConfiguration from "@/pages/things/ThingConfiguration.vue";

// Thing Details View
const props = defineProps<{
  td:ThingTD,
  height?:string,
}>()

const selectedTab = ref('attr')

</script>

<template>
  <div style="display: flex; flex-direction: column; overflow: auto; width: 100%; height: 100%">
    <QTabs horizontal align="left" v-model="selectedTab" style="color:brown">
      <QTab label="Attributes" :icon="matDescription" name="attr"/>
      <QTab label="Configuration" :icon="matSettings" name="config"/>
      <QTab label="Thing Events" :icon="matSettingsRemote" name="events"/>
      <QTab label="Thing Actions" :icon="matDirectionsRun" name="actions"/>
    </QTabs>

    <QTabPanels v-model="selectedTab">
      <QTabPanel name="attr"  class="q-pa-xs">
        <ThingAttributes :td="props.td"/>
      </QTabPanel>

      <QTabPanel name="config" class="q-pa-xs">
        <ThingConfiguration :td="props.td"/>
      </QTabPanel>

      <QTabPanel name="events" class="q-pa-xs">
        <ThingEvents :td="props.td"/>
      </QTabPanel>

      <QTabPanel name="actions" class="q-pa-xs">
        <ThingActions :td="props.td"/>
      </QTabPanel>

    </QTabPanels>
  </div>
</template>
