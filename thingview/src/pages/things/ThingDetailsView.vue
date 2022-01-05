<script lang="ts" setup>

import {ref} from 'vue'
import { date, QCard, QCardSection, QField, QForm, QTab, QTabs, QTabPanel, QTabPanels } from 'quasar';
const {formatDate}= date

import {matSettings, matSettingsRemote, matDescription, matDirectionsRun} from '@quasar/extras/material-icons'


import {ThingTD, TDProperty, TDAction, TDEvent} from '@/data/td/ThingTD';

import ThingEvents from './ThingEvents.vue'
import ThingActions from "@/pages/things/ThingActions.vue";
import ThingAttributes from "@/pages/things/ThingPropertiesTable.vue";
import ThingConfiguration from "@/pages/things/ThingConfiguration.vue";

// Thing Details View
const props = defineProps<{
  td:ThingTD,
  height?:string,
}>()

const selectedTab = ref('attr')

// Convert iso9601 date format to text representation 
const getDateText = (iso:string): string => {
  let timeStamp = new Date(iso)
  return formatDate(timeStamp, "ddd Do MMM YYYY HH:mm:ss (Z)")
}

</script>

<template>
  <div style="display: flex; flex-direction: column; overflow: auto; width: 100%; height: 100%">

  <QForm  class='row q-pb-sm'>
    <QField label="Thing ID" stack-label dense class="q-pl-md">
      {{props.td.id}}
    </QField>
    <QField  label="Created" stack-label dense class="q-pl-md">
      <!-- {{props.td.created}}  -->
      {{getDateText(props.td.created)}}
    </QField>
  </QForm>

    <QTabs horizontal align="left" v-model="selectedTab" >
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
